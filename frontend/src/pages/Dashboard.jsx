import React from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus, Loader2, Image as ImageIcon } from 'lucide-react';
import api from '../api/client';
import Layout from '../components/Layout';

const Dashboard = () => {
  const { data: user } = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const res = await api.get('/me');
      return res.data;
    },
  });

  const { data, isLoading, isError } = useQuery({
    queryKey: ['pixelarts'],
    queryFn: async () => {
      const response = await api.get('/pixelarts');
      return response.data;
    },
  });

  return (
    <Layout>
      <div className="container">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '2rem' }}>
          <div>
            <h1>My Gallery</h1>
            {user && <p style={{ color: 'var(--color-text-muted)' }}>Welcome, {user.username}!</p>}
          </div>
          <Link to="/editor" className="btn btn-primary">
            <Plus size={18} style={{ marginRight: '0.5rem' }} />
            New Art
          </Link>
        </div>

        {isLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem' }}>
            <Loader2 className="spin" size={32} color="var(--color-primary)" />
          </div>
        ) : isError ? (
          <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--color-error)' }}>
            Failed to load gallery.
          </div>
        ) : data?.previews?.length === 0 ? (
          <div className="empty-state">
            <div style={{ backgroundColor: 'var(--color-secondary)', padding: '1rem', borderRadius: '50%', marginBottom: '1rem' }}>
              <ImageIcon size={32} color="var(--color-text-muted)" />
            </div>
            <h3>No pixel arts yet</h3>
            <p style={{ color: 'var(--color-text-muted)', marginBottom: '1.5rem' }}>Start creating your first masterpiece!</p>
            <Link to="/editor" className="btn btn-primary">Create New</Link>
          </div>
        ) : (
          <div className="grid">
            {data?.previews.map((art) => (
              <Link key={art.pixelart_id} to={`/editor/${art.pixelart_id}`} className="art-card">
                <div className="art-image-container">
                  {art.image_url ? (
                    <img src={art.image_url} alt={art.title} className="art-image" />
                  ) : (
                    <div className="art-placeholder">
                      <ImageIcon size={24} color="var(--color-text-muted)" />
                    </div>
                  )}
                </div>
                <div className="art-info">
                  <h3 className="art-title">{art.title}</h3>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>

      <style>{`
        .spin {
          animation: spin 1s linear infinite;
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 4rem 2rem;
          background-color: var(--color-surface);
          border-radius: var(--radius-lg);
          border: 1px dashed var(--color-border);
        }
        .grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
          gap: var(--space-6);
        }
        .art-card {
          background-color: var(--color-surface);
          border-radius: var(--radius-lg);
          border: 1px solid var(--color-border);
          overflow: hidden;
          transition: transform 0.2s, box-shadow 0.2s;
        }
        .art-card:hover {
          transform: translateY(-2px);
          box-shadow: var(--shadow-md);
          border-color: var(--color-primary);
        }
        .art-image-container {
          aspect-ratio: 1;
          background-color: #ffffff;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        .art-image {
          width: 100%;
          height: 100%;
          object-fit: contain;
          image-rendering: pixelated;
        }
        .art-placeholder {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 100%;
          height: 100%;
          background-color: var(--color-bg);
        }
        .art-info {
          padding: var(--space-4);
        }
        .art-title {
          font-size: 1rem;
          font-weight: 600;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      `}</style>
    </Layout>
  );
};

export default Dashboard;
