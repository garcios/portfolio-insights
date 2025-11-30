import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Wallet, ArrowRight, Loader2 } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';

const AuthPage = () => {
    const navigate = useNavigate();
    const { login, isAuthenticated } = useAuth();
    const [isLoading, setIsLoading] = useState(false);

    // Redirect if already authenticated
    useEffect(() => {
        if (isAuthenticated) {
            navigate('/dashboard');
        }
    }, [isAuthenticated, navigate]);

    const handleOAuthLogin = async () => {
        setIsLoading(true);
        try {
            await login(); // This will redirect to Hydra
        } catch (error) {
            console.error('Login failed:', error);
            setIsLoading(false);
        }
    };

    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            padding: '24px'
        }}>
            <div style={{
                width: '100%',
                maxWidth: '440px',
                background: 'white',
                borderRadius: '16px',
                boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)',
                overflow: 'hidden'
            }}>
                {/* Header */}
                <div style={{
                    padding: '48px 32px',
                    textAlign: 'center',
                    background: 'linear-gradient(to bottom, #f8f9fa, white)'
                }}>
                    <div style={{
                        width: '64px',
                        height: '64px',
                        borderRadius: '16px',
                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        margin: '0 auto 24px',
                        boxShadow: '0 8px 24px rgba(102, 126, 234, 0.4)'
                    }}>
                        <Wallet size={36} color="white" />
                    </div>
                    <h1 style={{
                        fontSize: '2rem',
                        fontWeight: '700',
                        color: '#1a1a1a',
                        marginBottom: '12px'
                    }}>
                        Portfolio Insights
                    </h1>
                    <p style={{
                        color: '#666',
                        fontSize: '1rem',
                        lineHeight: '1.6'
                    }}>
                        Your wealth, visualized and optimized.
                    </p>
                </div>

                {/* Login Section */}
                <div style={{ padding: '32px' }}>
                    <p style={{
                        textAlign: 'center',
                        color: '#666',
                        marginBottom: '24px',
                        fontSize: '0.875rem'
                    }}>
                        Sign in with your account to continue
                    </p>

                    <button
                        onClick={handleOAuthLogin}
                        disabled={isLoading}
                        style={{
                            width: '100%',
                            padding: '14px',
                            borderRadius: '10px',
                            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                            color: 'white',
                            border: 'none',
                            fontWeight: '600',
                            fontSize: '1rem',
                            cursor: isLoading ? 'not-allowed' : 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            gap: '10px',
                            transition: 'transform 0.2s, box-shadow 0.2s',
                            opacity: isLoading ? 0.7 : 1,
                            boxShadow: '0 4px 12px rgba(102, 126, 234, 0.3)'
                        }}
                        onMouseOver={(e) => {
                            if (!isLoading) {
                                e.currentTarget.style.transform = 'translateY(-2px)';
                                e.currentTarget.style.boxShadow = '0 8px 20px rgba(102, 126, 234, 0.4)';
                            }
                        }}
                        onMouseOut={(e) => {
                            e.currentTarget.style.transform = 'translateY(0)';
                            e.currentTarget.style.boxShadow = '0 4px 12px rgba(102, 126, 234, 0.3)';
                        }}
                    >
                        {isLoading ? (
                            <>
                                <Loader2 size={20} className="animate-spin" />
                                Redirecting...
                            </>
                        ) : (
                            <>
                                Sign In with OAuth
                                <ArrowRight size={20} />
                            </>
                        )}
                    </button>

                    <div style={{
                        marginTop: '24px',
                        padding: '16px',
                        background: '#f8f9fa',
                        borderRadius: '8px',
                        border: '1px solid #e9ecef'
                    }}>
                        <p style={{
                            fontSize: '0.75rem',
                            color: '#666',
                            textAlign: 'center',
                            lineHeight: '1.5',
                            margin: 0
                        }}>
                            🔒 Secure authentication powered by Ory Hydra
                        </p>
                    </div>
                </div>

                {/* Footer */}
                <div style={{
                    padding: '20px 32px',
                    background: '#f8f9fa',
                    borderTop: '1px solid #e9ecef',
                    textAlign: 'center'
                }}>
                    <p style={{
                        fontSize: '0.75rem',
                        color: '#999',
                        margin: 0
                    }}>
                        Portfolio Insights © 2025
                    </p>
                </div>
            </div>
            <style>{`
                .animate-spin {
                    animation: spin 1s linear infinite;
                }
                @keyframes spin {
                    from { transform: rotate(0deg); }
                    to { transform: rotate(360deg); }
                }
            `}</style>
        </div>
    );
};

export default AuthPage;
