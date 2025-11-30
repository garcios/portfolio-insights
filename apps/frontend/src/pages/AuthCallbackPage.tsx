import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

const AuthCallbackPage = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const { handleCallback } = useAuth();
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const processCallback = async () => {
            const code = searchParams.get('code');
            const state = searchParams.get('state');
            const errorParam = searchParams.get('error');
            const errorDescription = searchParams.get('error_description');

            if (errorParam) {
                setError(errorDescription || errorParam);
                return;
            }

            if (!code || !state) {
                setError('Missing authorization code or state');
                return;
            }

            try {
                await handleCallback(code, state);
                // Redirect to home page after successful authentication
                navigate('/', { replace: true });
            } catch (err) {
                console.error('Callback error:', err);
                setError(err instanceof Error ? err.message : 'Authentication failed');
            }
        };

        processCallback();
    }, [searchParams, handleCallback, navigate]);

    if (error) {
        return (
            <div style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: '100vh',
                background: 'var(--color-bg-primary)',
                padding: '20px'
            }}>
                <div style={{
                    background: 'var(--color-bg-card)',
                    borderRadius: '12px',
                    padding: '40px',
                    maxWidth: '500px',
                    width: '100%',
                    textAlign: 'center',
                    boxShadow: 'var(--shadow-lg)'
                }}>
                    <div style={{
                        fontSize: '48px',
                        marginBottom: '20px'
                    }}>⚠️</div>
                    <h1 style={{
                        color: 'var(--color-text-primary)',
                        fontSize: '24px',
                        marginBottom: '16px'
                    }}>
                        Authentication Error
                    </h1>
                    <p style={{
                        color: 'var(--color-text-secondary)',
                        marginBottom: '24px'
                    }}>
                        {error}
                    </p>
                    <button
                        onClick={() => navigate('/login')}
                        style={{
                            padding: '12px 24px',
                            background: 'var(--color-primary)',
                            color: 'white',
                            border: 'none',
                            borderRadius: '8px',
                            fontSize: '16px',
                            fontWeight: '600',
                            cursor: 'pointer',
                            transition: 'transform 0.2s'
                        }}
                    >
                        Try Again
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '100vh',
            background: 'var(--color-bg-primary)'
        }}>
            <div style={{
                textAlign: 'center',
                color: 'var(--color-text-primary)'
            }}>
                <div style={{
                    width: '48px',
                    height: '48px',
                    border: '4px solid var(--color-border)',
                    borderTopColor: 'var(--color-primary)',
                    borderRadius: '50%',
                    animation: 'spin 1s linear infinite',
                    margin: '0 auto 16px'
                }}></div>
                <p>Completing authentication...</p>
            </div>
        </div>
    );
};

export default AuthCallbackPage;
