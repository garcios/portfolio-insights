import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Wallet, Eye, EyeOff, ArrowRight, Loader2 } from 'lucide-react';
import Input from '../components/ui/Input';

const AuthPage = () => {
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = useState<'login' | 'register'>('login');
    const [isLoading, setIsLoading] = useState(false);
    const [showPassword, setShowPassword] = useState(false);

    // Login State
    const [loginData, setLoginData] = useState({ email: '', password: '', rememberMe: false });
    const [loginErrors, setLoginErrors] = useState<{ email?: string; password?: string }>({});

    // Register State
    const [registerData, setRegisterData] = useState({ name: '', email: '', password: '', confirmPassword: '' });
    const [registerErrors, setRegisterErrors] = useState<{ name?: string; email?: string; password?: string; confirmPassword?: string }>({});

    const validateEmail = (email: string) => {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    };

    const handleLoginSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const errors: typeof loginErrors = {};

        if (!loginData.email) errors.email = 'Email is required';
        else if (!validateEmail(loginData.email)) errors.email = 'Invalid email format';

        if (!loginData.password) errors.password = 'Password is required';

        setLoginErrors(errors);

        if (Object.keys(errors).length === 0) {
            setIsLoading(true);
            // Simulate API call
            setTimeout(() => {
                setIsLoading(false);
                navigate('/');
            }, 1500);
        }
    };

    const handleRegisterSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const errors: typeof registerErrors = {};

        if (!registerData.name) errors.name = 'Name is required';

        if (!registerData.email) errors.email = 'Email is required';
        else if (!validateEmail(registerData.email)) errors.email = 'Invalid email format';

        if (!registerData.password) errors.password = 'Password is required';
        else if (registerData.password.length < 8) errors.password = 'Password must be at least 8 characters';

        if (registerData.password !== registerData.confirmPassword) {
            errors.confirmPassword = 'Passwords do not match';
        }

        setRegisterErrors(errors);

        if (Object.keys(errors).length === 0) {
            setIsLoading(true);
            // Simulate API call
            setTimeout(() => {
                setIsLoading(false);
                // Auto login or switch to login tab
                setActiveTab('login');
                setLoginData(prev => ({ ...prev, email: registerData.email }));
                alert('Registration successful! Please log in.');
            }, 1500);
        }
    };

    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'var(--color-bg-secondary)',
            padding: '24px'
        }}>
            <div style={{
                width: '100%',
                maxWidth: '440px',
                background: 'var(--color-bg-card)',
                borderRadius: '16px',
                boxShadow: 'var(--shadow-xl)',
                border: '1px solid var(--color-border)',
                overflow: 'hidden'
            }}>
                {/* Header */}
                <div style={{
                    padding: '32px 32px 24px',
                    textAlign: 'center',
                    background: 'linear-gradient(to bottom, var(--color-bg-tertiary), var(--color-bg-card))'
                }}>
                    <div style={{
                        width: '48px',
                        height: '48px',
                        borderRadius: '12px',
                        background: 'linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        margin: '0 auto 16px',
                        boxShadow: '0 4px 12px rgba(99, 102, 241, 0.3)'
                    }}>
                        <Wallet size={28} color="white" />
                    </div>
                    <h1 style={{
                        fontSize: '1.5rem',
                        fontWeight: '700',
                        color: 'var(--color-text-primary)',
                        marginBottom: '8px'
                    }}>
                        Portfolio Insights
                    </h1>
                    <p style={{ color: 'var(--color-text-tertiary)', fontSize: '0.875rem' }}>
                        Your wealth, visualized and optimized.
                    </p>
                </div>

                {/* Tabs */}
                <div style={{
                    display: 'flex',
                    borderBottom: '1px solid var(--color-border)',
                    padding: '0 16px'
                }}>
                    <button
                        onClick={() => setActiveTab('login')}
                        style={{
                            flex: 1,
                            padding: '16px',
                            background: 'transparent',
                            border: 'none',
                            borderBottom: activeTab === 'login' ? '2px solid var(--color-primary)' : '2px solid transparent',
                            color: activeTab === 'login' ? 'var(--color-primary)' : 'var(--color-text-secondary)',
                            fontWeight: '600',
                            cursor: 'pointer',
                            transition: 'all 0.2s'
                        }}
                    >
                        Login
                    </button>
                    <button
                        onClick={() => setActiveTab('register')}
                        style={{
                            flex: 1,
                            padding: '16px',
                            background: 'transparent',
                            border: 'none',
                            borderBottom: activeTab === 'register' ? '2px solid var(--color-primary)' : '2px solid transparent',
                            color: activeTab === 'register' ? 'var(--color-primary)' : 'var(--color-text-secondary)',
                            fontWeight: '600',
                            cursor: 'pointer',
                            transition: 'all 0.2s'
                        }}
                    >
                        Register
                    </button>
                </div>

                {/* Forms */}
                <div style={{ padding: '32px' }}>
                    {activeTab === 'login' ? (
                        <form onSubmit={handleLoginSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                            <Input
                                id="login-email"
                                label="Email Address"
                                type="email"
                                placeholder="john@example.com"
                                value={loginData.email}
                                onChange={(e) => setLoginData({ ...loginData, email: e.target.value })}
                                error={loginErrors.email}
                            />

                            <div style={{ position: 'relative' }}>
                                <Input
                                    id="login-password"
                                    label="Password"
                                    type={showPassword ? "text" : "password"}
                                    placeholder="••••••••"
                                    value={loginData.password}
                                    onChange={(e) => setLoginData({ ...loginData, password: e.target.value })}
                                    error={loginErrors.password}
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    style={{
                                        position: 'absolute',
                                        right: '12px',
                                        top: '32px',
                                        background: 'transparent',
                                        border: 'none',
                                        color: 'var(--color-text-tertiary)',
                                        cursor: 'pointer'
                                    }}
                                >
                                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                                </button>
                            </div>

                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
                                    <input
                                        type="checkbox"
                                        checked={loginData.rememberMe}
                                        onChange={(e) => setLoginData({ ...loginData, rememberMe: e.target.checked })}
                                        style={{ accentColor: 'var(--color-primary)' }}
                                    />
                                    <span style={{ fontSize: '0.875rem', color: 'var(--color-text-secondary)' }}>Remember me</span>
                                </label>
                                <a href="#" style={{ fontSize: '0.875rem', color: 'var(--color-primary)', textDecoration: 'none', fontWeight: '500' }}>
                                    Forgot password?
                                </a>
                            </div>

                            <button
                                type="submit"
                                disabled={isLoading}
                                style={{
                                    marginTop: '8px',
                                    padding: '12px',
                                    borderRadius: '8px',
                                    background: 'var(--color-primary)',
                                    color: 'white',
                                    border: 'none',
                                    fontWeight: '600',
                                    fontSize: '1rem',
                                    cursor: isLoading ? 'not-allowed' : 'pointer',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    gap: '8px',
                                    transition: 'background 0.2s',
                                    opacity: isLoading ? 0.7 : 1
                                }}
                            >
                                {isLoading ? <Loader2 size={20} className="animate-spin" /> : <>Sign In <ArrowRight size={18} /></>}
                            </button>
                        </form>
                    ) : (
                        <form onSubmit={handleRegisterSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                            <Input
                                id="register-name"
                                label="Full Name"
                                type="text"
                                placeholder="John Doe"
                                value={registerData.name}
                                onChange={(e) => setRegisterData({ ...registerData, name: e.target.value })}
                                error={registerErrors.name}
                            />

                            <Input
                                id="register-email"
                                label="Email Address"
                                type="email"
                                placeholder="john@example.com"
                                value={registerData.email}
                                onChange={(e) => setRegisterData({ ...registerData, email: e.target.value })}
                                error={registerErrors.email}
                            />

                            <Input
                                id="register-password"
                                label="Password"
                                type="password"
                                placeholder="••••••••"
                                value={registerData.password}
                                onChange={(e) => setRegisterData({ ...registerData, password: e.target.value })}
                                error={registerErrors.password}
                            />

                            <Input
                                id="register-confirm-password"
                                label="Confirm Password"
                                type="password"
                                placeholder="••••••••"
                                value={registerData.confirmPassword}
                                onChange={(e) => setRegisterData({ ...registerData, confirmPassword: e.target.value })}
                                error={registerErrors.confirmPassword}
                            />

                            <div style={{ fontSize: '0.75rem', color: 'var(--color-text-tertiary)', lineHeight: '1.4' }}>
                                Password must be at least 8 characters long and contain a mix of letters and numbers.
                            </div>

                            <button
                                type="submit"
                                disabled={isLoading}
                                style={{
                                    marginTop: '8px',
                                    padding: '12px',
                                    borderRadius: '8px',
                                    background: 'var(--color-primary)',
                                    color: 'white',
                                    border: 'none',
                                    fontWeight: '600',
                                    fontSize: '1rem',
                                    cursor: isLoading ? 'not-allowed' : 'pointer',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    gap: '8px',
                                    transition: 'background 0.2s',
                                    opacity: isLoading ? 0.7 : 1
                                }}
                            >
                                {isLoading ? <Loader2 size={20} className="animate-spin" /> : 'Create Account'}
                            </button>
                        </form>
                    )}
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
