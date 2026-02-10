import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Save, ArrowLeft, Trash2, Eraser, Pencil, Check } from 'lucide-react';
import api from '../api/client';
import Layout from '../components/Layout';
import clsx from 'clsx';

const DEFAULT_PALETTE = [
    '#000000', '#1d2b53', '#7e2553', '#008751', '#ab5236', '#5f574f', '#c2c3c7', '#fff1e8',
    '#ff004d', '#ffa300', '#ffec27', '#00e436', '#29adff', '#83769c', '#ff77a8', '#ffccaa',
    '#ffffff', '#99e550', '#6abe30', '#37946e', '#4b692f', '#524b24', '#323c39', '#3f3f74'
];

const DEFAULT_BG_INDEX = 16;

const Editor = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    const [isSetup, setIsSetup] = useState(!id);
    const [setupTitle, setSetupTitle] = useState('Untitled Art');
    const [setupWidth, setSetupWidth] = useState(32);
    const [setupHeight, setSetupHeight] = useState(32);

    const [title, setTitle] = useState('Untitled Art');
    const [width, setWidth] = useState(16);
    const [height, setHeight] = useState(16);

    const [palette, setPalette] = useState(DEFAULT_PALETTE);

    const pixelsRef = useRef([]);

    const [selectedColor, setSelectedColor] = useState(0);
    const [tool, setTool] = useState('pencil');

    const canvasRef = useRef(null);
    const isDrawing = useRef(false);
    const lastPixel = useRef(null);

    const { data, isLoading } = useQuery({
        queryKey: ['pixelart', id],
        queryFn: async () => {
            const res = await api.get(`/pixelart/${id}`);
            return res.data;
        },
        enabled: !!id,
    });

    useEffect(() => {
        if (data) {
            setTitle(data.title);
            setWidth(data.width);
            setHeight(data.height);
            setPalette(data.palette);
            pixelsRef.current = data.pixels;
            setIsSetup(false);
            drawCanvas();
        }
    }, [data]);

    useEffect(() => {
        if (!isSetup && canvasRef.current) {
            drawCanvas();
        }
    }, [width, height, isSetup]);

    const drawCanvas = () => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        const p = pixelsRef.current;

        ctx.clearRect(0, 0, width, height);

        for (let y = 0; y < height; y++) {
            for (let x = 0; x < width; x++) {
                const idx = y * width + x;
                const colorIdx = p[idx] !== undefined ? p[idx] : DEFAULT_BG_INDEX;
                const colorStyle = palette[colorIdx] || '#ffffff';
                ctx.fillStyle = colorStyle;
                ctx.fillRect(x, y, 1, 1);
            }
        }
    };

    const drawPixel = (x, y, colorIdx) => {
        const idx = y * width + x;
        if (x < 0 || x >= width || y < 0 || y >= height) return;

        pixelsRef.current[idx] = colorIdx;

        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        ctx.fillStyle = palette[colorIdx];
        ctx.fillRect(x, y, 1, 1);
    };

    const handleSetupSubmit = (e) => {
        e.preventDefault();
        setTitle(setupTitle);
        setWidth(setupWidth);
        setHeight(setupHeight);
        pixelsRef.current = Array(setupWidth * setupHeight).fill(DEFAULT_BG_INDEX);
        setIsSetup(false);
    };

    const saveMutation = useMutation({
        mutationFn: async (data) => {
            if (id) {
                await api.patch(`/pixelart/${id}`, {
                    palette: data.palette,
                    pixels: data.pixels,
                    width: data.width,
                    height: data.height
                });
                return { pixelart_id: id };
            } else {
                const res = await api.post('/pixelart', data);
                return res.data;
            }
        },
        onSuccess: (data) => {
            queryClient.invalidateQueries(['pixelarts']);
            if (!id) {
                navigate(`/editor/${data.pixelart_id}`);
            }
        },
    });

    const deleteMutation = useMutation({
        mutationFn: async () => {
            await api.delete(`/pixelart/${id}`);
        },
        onSuccess: () => {
            queryClient.invalidateQueries(['pixelarts']);
            navigate('/dashboard');
        },
    });

    const handleSave = () => {
        saveMutation.mutate({
            title,
            palette,
            pixels: pixelsRef.current,
            width,
            height
        });
    };

    const drawLine = (x0, y0, x1, y1) => {
        const dx = Math.abs(x1 - x0);
        const dy = Math.abs(y1 - y0);
        const sx = (x0 < x1) ? 1 : -1;
        const sy = (y0 < y1) ? 1 : -1;
        let err = dx - dy;

        let x = x0;
        let y = y0;

        while (true) {
            drawPixel(x, y, tool === 'eraser' ? DEFAULT_BG_INDEX : selectedColor);

            if (x === x1 && y === y1) break;
            const e2 = 2 * err;
            if (e2 > -dy) { err -= dy; x += sx; }
            if (e2 < dx) { err += dx; y += sy; }
        }
    };

    const getPointerPos = (e) => {
        const canvas = canvasRef.current;
        if (!canvas) return { x: 0, y: 0 };
        const rect = canvas.getBoundingClientRect();
        const x = Math.floor((e.clientX - rect.left) / (rect.width / width));
        const y = Math.floor((e.clientY - rect.top) / (rect.height / height));
        return { x, y };
    };

    const handlePointerDown = (e) => {
        e.preventDefault();
        isDrawing.current = true;
        const { x, y } = getPointerPos(e);
        lastPixel.current = { x, y };

        drawPixel(x, y, tool === 'eraser' ? DEFAULT_BG_INDEX : selectedColor);
    };

    const handlePointerMove = (e) => {
        if (!isDrawing.current) return;

        const { x, y } = getPointerPos(e);

        if (lastPixel.current) {
            drawLine(lastPixel.current.x, lastPixel.current.y, x, y);
        }
        lastPixel.current = { x, y };
    };

    const handlePointerUp = () => {
        isDrawing.current = false;
        lastPixel.current = null;
    };

    useEffect(() => {
        window.addEventListener('pointerup', handlePointerUp);
        window.addEventListener('pointercancel', handlePointerUp);
        return () => {
            window.removeEventListener('pointerup', handlePointerUp);
            window.removeEventListener('pointercancel', handlePointerUp);
        };
    }, []);

    if (isLoading && id) return <Layout><div>Loading...</div></Layout>;

    if (isSetup) {
        return (
            <Layout>
                <div className="container" style={{ maxWidth: '400px', marginTop: '4rem' }}>
                    <div className="card">
                        <h1 style={{ marginBottom: '1.5rem' }}>New Pixel Art</h1>
                        <form onSubmit={handleSetupSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                            <div>
                                <label className="label">Title</label>
                                <input
                                    className="input"
                                    value={setupTitle}
                                    onChange={e => setSetupTitle(e.target.value)}
                                    required
                                />
                            </div>
                            <div style={{ display: 'flex', gap: '1rem' }}>
                                <div style={{ flex: 1 }}>
                                    <label className="label">Width (8-64)</label>
                                    <input
                                        type="number"
                                        className="input"
                                        value={setupWidth}
                                        onChange={e => setSetupWidth(Number(e.target.value))}
                                        min={8}
                                        max={64}
                                        required
                                    />
                                </div>
                                <div style={{ flex: 1 }}>
                                    <label className="label">Height (8-64)</label>
                                    <input
                                        type="number"
                                        className="input"
                                        value={setupHeight}
                                        onChange={e => setSetupHeight(Number(e.target.value))}
                                        min={8}
                                        max={64}
                                        required
                                    />
                                </div>
                            </div>
                            <button type="submit" className="btn btn-primary" style={{ marginTop: '1rem' }}>
                                Create Canvas
                            </button>
                        </form>
                    </div>
                </div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="container editor-container">
                <div className="toolbar">
                    <div className="left">
                        <button className="btn btn-secondary" onClick={() => navigate('/dashboard')}>
                            <ArrowLeft size={16} style={{ marginRight: '0.5rem' }} /> Back
                        </button>
                        <input
                            type="text"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            className="title-input"
                            disabled={!!id}
                        />
                    </div>
                    <div className="right">
                        {id && (
                            <button className="btn btn-danger" onClick={() => deleteMutation.mutate()} style={{ marginRight: '0.5rem' }}>
                                <Trash2 size={16} />
                            </button>
                        )}
                        <button className="btn btn-primary" onClick={handleSave} disabled={saveMutation.isPending}>
                            <Save size={16} style={{ marginRight: '0.5rem' }} />
                            {saveMutation.isPending ? 'Saving...' : 'Save'}
                        </button>
                    </div>
                </div>

                <div className="workspace">
                    <div className="tools-panel">
                        <div className="panel-section">
                            <h3>Tools</h3>
                            <div className="tools-grid">
                                <button
                                    className={clsx('tool-btn', tool === 'pencil' && 'active')}
                                    onClick={() => setTool('pencil')}
                                >
                                    <Pencil size={20} />
                                </button>
                                <button
                                    className={clsx('tool-btn', tool === 'eraser' && 'active')}
                                    onClick={() => setTool('eraser')}
                                >
                                    <Eraser size={20} />
                                </button>
                            </div>
                        </div>

                        <div className="panel-section">
                            <h3>Palette</h3>
                            <div className="palette-grid">
                                {palette.map((color, idx) => (
                                    <button
                                        key={idx}
                                        className={clsx('color-btn', selectedColor === idx && 'active')}
                                        style={{ backgroundColor: color }}
                                        onClick={() => {
                                            setSelectedColor(idx);
                                            setTool('pencil');
                                        }}
                                    >
                                        {selectedColor === idx && <div className="color-check"><Check size={12} strokeWidth={4} /></div>}
                                    </button>
                                ))}
                            </div>
                        </div>
                    </div>

                    <div className="canvas-wrapper">
                        <canvas
                            ref={canvasRef}
                            width={width}
                            height={height}
                            className="pixel-canvas"
                            style={{
                                width: 'min(500px, 90vw)',
                                aspectRatio: `${width}/${height}`,
                            }}
                            onPointerDown={handlePointerDown}
                            onPointerMove={handlePointerMove}
                        />
                    </div>
                </div>
            </div>

            <style>{`
        .editor-container {
          display: flex;
          flex-direction: column;
          gap: var(--space-4);
        }
        .toolbar {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: var(--space-4);
          background-color: var(--color-surface);
          border-radius: var(--radius-lg);
          border: 1px solid var(--color-border);
        }
        .left, .right {
          display: flex;
          align-items: center;
          gap: var(--space-4);
        }
        .title-input {
          font-size: 1.25rem;
          font-weight: 600;
          border: none;
          background: transparent;
          color: var(--color-text);
        }
        .title-input:focus {
          outline: none;
          text-decoration: underline;
        }
        
        .workspace {
          display: flex;
          gap: var(--space-6);
        }
        
        .tools-panel {
          width: 250px;
          display: flex;
          flex-direction: column;
          gap: var(--space-4);
          flex-shrink: 0;
        }
        .panel-section {
          background-color: var(--color-surface);
          border-radius: var(--radius-lg);
          border: 1px solid var(--color-border);
          padding: var(--space-4);
        }
        .panel-section h3 {
          font-size: 0.875rem;
          text-transform: uppercase;
          letter-spacing: 0.05em;
          color: var(--color-text-muted);
          margin-bottom: var(--space-4);
        }

        .tools-grid {
          display: flex;
          gap: var(--space-2);
        }
        .tool-btn {
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: var(--radius-md);
          border: 1px solid var(--color-border);
          color: var(--color-text-muted);
          background-color: var(--color-bg);
          transition: all 0.1s;
        }
        .tool-btn.active {
          background-color: var(--color-primary);
          color: white;
          border-color: var(--color-primary);
        }

        .palette-grid {
          display: grid;
          grid-template-columns: repeat(4, 1fr);
          gap: var(--space-2);
        }
        .color-btn {
          width: 100%;
          aspect-ratio: 1;
          border-radius: var(--radius-sm);
          border: 1px solid rgba(0,0,0,0.1);
          position: relative;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        .color-btn.active {
          transform: scale(1.1);
          box-shadow: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-primary);
          z-index: 10;
        }
        .color-check {
            color: #333; 
            filter: drop-shadow(0 0 2px rgba(255,255,255,0.8));
        }

        .canvas-wrapper {
          flex: 1;
          background-color: #ffffff;
          border-radius: var(--radius-lg);
          border: 1px solid var(--color-border);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 2rem;
          min-height: 500px;
        }

        .pixel-canvas {
          background-color: white;
          box-shadow: var(--shadow-lg);
          image-rendering: pixelated; 
          cursor: crosshair;
          touch-action: none;
          
          border: 1px solid #cbd5e1;
        }

        @media (max-width: 768px) {
          .workspace {
            flex-direction: column;
          }
          .tools-panel {
            width: 100%;
          }
          .palette-grid {
             grid-template-columns: repeat(8, 1fr);
          }
        }
        
        .label { display: block; margin-bottom: 0.5rem; font-weight: 500; font-size: 0.875rem; }
      `}</style>
        </Layout>
    );
};

export default Editor;
