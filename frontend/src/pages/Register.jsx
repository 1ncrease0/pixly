import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import api from '../api/client';
import Layout from '../components/Layout';
import { Loader2 } from 'lucide-react';

const Register = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({ email: '', username: '', password: '' });
    const [error, setError] = useState('');

    const registerMutation = useMutation({
        mutationFn: async (data) => {
            const response = await api.post('/auth/register', data);
            return response.data;
        },
        onSuccess: () => {
            navigate('/login');
        },
        onError: (err) => {
            setError(err.response?.data?.error?.message || 'Failed to register');
        },
    });

    const handleSubmit = (e) => {
        e.preventDefault();
        setError('');
        registerMutation.mutate(formData);
    };

    return (
        <Layout>
            <div className="container" style={{ maxWidth: '400px' }}>
                <div className="card">
                    <h1 style={{ marginBottom: '1.5rem', textAlign: 'center' }}>Create Account</h1>

                    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                        {error && (
                            <div style={{
                                padding: '0.75rem',
                                backgroundColor: '#fee2e2',
                                color: '#ef4444',
                                borderRadius: 'var(--radius-md)',
                                fontSize: '0.875rem'
                            }}>
                                {error}
                            </div>
                        )}

                        <div>
                            <label htmlFor="username" style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.875rem', fontWeight: 500 }}>Username</label>
                            <input
                                id="username"
                                type="text"
                                className="input"
                                value={formData.username}
                                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                                required
                                placeholder="PixelArtist"
                            />
                        </div>

                        <div>
                            <label htmlFor="email" style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.875rem', fontWeight: 500 }}>Email</label>
                            <input
                                id="email"
                                type="email"
                                className="input"
                                value={formData.email}
                                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                required
                                placeholder="you@example.com"
                            />
                        </div>

                        <div>
                            <label htmlFor="password" style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.875rem', fontWeight: 500 }}>Password</label>
                            <input
                                id="password"
                                type="password"
                                className="input"
                                value={formData.password}
                                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                                required
                                minLength={8}
                                placeholder="••••••••"
                            />
                        </div>

                        <button
                            type="submit"
                            className="btn btn-primary"
                            style={{ marginTop: '0.5rem' }}
                            disabled={registerMutation.isPending}
                        >
                            {registerMutation.isPending ? (
                                <>
                                    <Loader2 className="spin" size={18} style={{ marginRight: '0.5rem' }} />
                                    Creating account...
                                </>
                            ) : 'Sign Up'}
                        </button>
                    </form>

                    <p style={{ marginTop: '1.5rem', textAlign: 'center', fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                        Already have an account? <Link to="/login" style={{ color: 'var(--color-primary)', fontWeight: 500 }}>Sign in</Link>
                    </p>
                </div>
            </div>
            <style>{`
        .spin {
          animation: spin 1s linear infinite;
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
        </Layout>
    );
};

export default Register;
