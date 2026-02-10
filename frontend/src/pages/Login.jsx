import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import api from '../api/client';
import Layout from '../components/Layout';
import { Loader2 } from 'lucide-react';

const Login = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({ email: '', password: '' });
    const [error, setError] = useState('');

    const loginMutation = useMutation({
        mutationFn: async (data) => {
            const response = await api.post('/auth/login', data);
            return response.data;
        },
        onSuccess: (data) => {
            localStorage.setItem('access_token', data.access_token);
            localStorage.setItem('refresh_token', data.refresh_token);
            api.defaults.headers.Authorization = `Bearer ${data.access_token}`;
            navigate('/dashboard');
        },
        onError: (err) => {
            setError(err.response?.data?.error?.message || 'Failed to login');
        },
    });

    const resendMutation = useMutation({
        mutationFn: async (email) => {
            const response = await api.post('/auth/resend-verification', { email });
            return response.data;
        },
        onSuccess: () => {
            setError('');
            alert('Verification email sent! Please check your inbox.');
        },
        onError: (err) => {
            alert(err.response?.data?.error?.message || 'Failed to resend email');
        }
    });

    const handleSubmit = (e) => {
        e.preventDefault();
        setError('');
        loginMutation.mutate(formData);
    };

    return (
        <Layout>
            <div className="container" style={{ maxWidth: '400px' }}>
                <div className="card">
                    <h1 style={{ marginBottom: '1.5rem', textAlign: 'center' }}>Welcome Back</h1>

                    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                        {error && (
                            <div style={{
                                padding: '0.75rem',
                                backgroundColor: '#fee2e2',
                                color: '#ef4444',
                                borderRadius: 'var(--radius-md)',
                                fontSize: '0.875rem',
                                display: 'flex',
                                flexDirection: 'column',
                                gap: '0.5rem'
                            }}>
                                <span>{error}</span>
                                {error.toLowerCase().includes('verified') && (
                                    <button
                                        type="button"
                                        onClick={() => resendMutation.mutate(formData.email)}
                                        className="btn"
                                        style={{ backgroundColor: 'white', color: '#ef4444', padding: '0.25rem 0.5rem', fontSize: '0.75rem', width: 'fit-content' }}
                                    >
                                        {resendMutation.isPending ? 'Sending...' : 'Resend Verification Email'}
                                    </button>
                                )}
                            </div>
                        )}

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
                                placeholder="••••••••"
                            />
                        </div>

                        <button
                            type="submit"
                            className="btn btn-primary"
                            style={{ marginTop: '0.5rem' }}
                            disabled={loginMutation.isPending}
                        >
                            {loginMutation.isPending ? (
                                <>
                                    <Loader2 className="spin" size={18} style={{ marginRight: '0.5rem' }} />
                                    Signing in...
                                </>
                            ) : 'Sign In'}
                        </button>
                    </form>

                    <p style={{ marginTop: '1.5rem', textAlign: 'center', fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                        Don't have an account? <Link to="/register" style={{ color: 'var(--color-primary)', fontWeight: 500 }}>Sign up</Link>
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

export default Login;
