import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { Loader2, CheckCircle, XCircle } from 'lucide-react';
import api from '../api/client';
import Layout from '../components/Layout';

const VerifyEmail = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const token = searchParams.get('token');
    const [status, setStatus] = useState('verifying');
    const [message, setMessage] = useState('');

    const verifyMutation = useMutation({
        mutationFn: async (token) => {
            const response = await api.post(`/auth/verify?token=${token}`);
            return response.data;
        },
        onSuccess: (data) => {
            setStatus('success');
            setMessage(data.message || 'Email verified successfully!');
            setTimeout(() => {
                navigate('/login');
            }, 3000);
        },
        onError: (err) => {
            setStatus('error');
            setMessage(err.response?.data?.error?.message || 'Failed to verify email. The link might be expired or invalid.');
        },
    });

    useEffect(() => {
        if (token) {
            verifyMutation.mutate(token);
        } else {
            setStatus('error');
            setMessage('Invalid verification link.');
        }
    }, [token]);

    return (
        <Layout>
            <div className="container" style={{ maxWidth: '500px', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', paddingTop: '4rem' }}>
                <div className="card" style={{ width: '100%', padding: '3rem 2rem' }}>

                    {status === 'verifying' && (
                        <>
                            <Loader2 className="spin" size={48} color="var(--color-primary)" style={{ marginBottom: '1.5rem' }} />
                            <h2>Verifying your email...</h2>
                            <p style={{ color: 'var(--color-text-muted)', marginTop: '0.5rem' }}>Please wait a moment.</p>
                        </>
                    )}

                    {status === 'success' && (
                        <>
                            <CheckCircle size={48} color="var(--color-success)" style={{ marginBottom: '1.5rem' }} />
                            <h2>Verified!</h2>
                            <p style={{ color: 'var(--color-text-muted)', marginTop: '0.5rem', marginBottom: '1.5rem' }}>{message}</p>
                            <p style={{ fontSize: '0.875rem' }}>Redirecting to login...</p>
                            <Link to="/login" className="btn btn-primary" style={{ marginTop: '1rem' }}>Go to Login</Link>
                        </>
                    )}

                    {status === 'error' && (
                        <>
                            <XCircle size={48} color="var(--color-error)" style={{ marginBottom: '1.5rem' }} />
                            <h2>Verification Failed</h2>
                            <p style={{ color: 'var(--color-text-muted)', marginTop: '0.5rem', marginBottom: '1.5rem' }}>{message}</p>
                            <Link to="/login" className="btn btn-secondary">Back to Login</Link>
                        </>
                    )}

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

export default VerifyEmail;
