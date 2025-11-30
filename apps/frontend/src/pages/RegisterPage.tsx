import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, gql } from '@apollo/client';
import {
    User,
    Mail,
    Lock,
    ArrowRight,
    CheckCircle2,
    AlertCircle,
    Loader2,
    Sparkles
} from 'lucide-react';
import './RegisterPage.css';

const CREATE_USER = gql`
    mutation CreateUser($input: NewUser!) {
        createUser(input: $input) {
            id
            username
            email
        }
    }
`;

interface FormData {
    username: string;
    email: string;
    password: string;
    confirmPassword: string;
}

interface FormErrors {
    username?: string;
    email?: string;
    password?: string;
    confirmPassword?: string;
    general?: string;
}

const RegisterPage = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState<FormData>({
        username: '',
        email: '',
        password: '',
        confirmPassword: ''
    });
    const [errors, setErrors] = useState<FormErrors>({});
    const [showSuccess, setShowSuccess] = useState(false);

    const [createUser, { loading }] = useMutation(CREATE_USER, {
        onCompleted: () => {
            setShowSuccess(true);
            setTimeout(() => {
                navigate('/login');
            }, 2000);
        },
        onError: (error) => {
            setErrors({ general: error.message });
        }
    });

    const validateForm = (): boolean => {
        const newErrors: FormErrors = {};

        // Username validation
        if (!formData.username.trim()) {
            newErrors.username = 'Username is required';
        } else if (formData.username.length < 3) {
            newErrors.username = 'Username must be at least 3 characters';
        } else if (!/^[a-zA-Z0-9_]+$/.test(formData.username)) {
            newErrors.username = 'Username can only contain letters, numbers, and underscores';
        }

        // Email validation
        if (!formData.email.trim()) {
            newErrors.email = 'Email is required';
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
            newErrors.email = 'Please enter a valid email address';
        }

        // Password validation
        if (!formData.password) {
            newErrors.password = 'Password is required';
        } else if (formData.password.length < 8) {
            newErrors.password = 'Password must be at least 8 characters';
        } else if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(formData.password)) {
            newErrors.password = 'Password must contain uppercase, lowercase, and number';
        }

        // Confirm password validation
        if (!formData.confirmPassword) {
            newErrors.confirmPassword = 'Please confirm your password';
        } else if (formData.password !== formData.confirmPassword) {
            newErrors.confirmPassword = 'Passwords do not match';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!validateForm()) {
            return;
        }

        try {
            await createUser({
                variables: {
                    input: {
                        username: formData.username,
                        email: formData.email,
                        password: formData.password
                    }
                }
            });
        } catch (error) {
            // Error handled by onError callback
            console.error('Registration error:', error);
        }
    };

    const handleChange = (field: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData(prev => ({ ...prev, [field]: e.target.value }));
        // Clear error for this field when user starts typing
        if (errors[field]) {
            setErrors(prev => ({ ...prev, [field]: undefined }));
        }
    };

    if (showSuccess) {
        return (
            <div className="register-page">
                <div className="register-background">
                    <div className="gradient-orb orb-1"></div>
                    <div className="gradient-orb orb-2"></div>
                </div>

                <div className="success-container">
                    <div className="success-icon">
                        <CheckCircle2 size={64} />
                    </div>
                    <h1 className="success-title">Account Created!</h1>
                    <p className="success-message">
                        Welcome to Portfolio Insights. Redirecting you to sign in...
                    </p>
                    <div className="spinner"></div>
                </div>
            </div>
        );
    }

    return (
        <div className="register-page">
            <div className="register-background">
                <div className="gradient-orb orb-1"></div>
                <div className="gradient-orb orb-2"></div>
                <div className="gradient-orb orb-3"></div>
            </div>

            <div className="register-container">
                <div className="register-card">
                    {/* Header */}
                    <div className="register-header">
                        <div className="brand-badge">
                            <Sparkles size={20} />
                            <span>Portfolio Insights</span>
                        </div>
                        <h1 className="register-title">Create Your Account</h1>
                        <p className="register-subtitle">
                            Start your journey to smarter investing
                        </p>
                    </div>

                    {/* Form */}
                    <form onSubmit={handleSubmit} className="register-form">
                        {errors.general && (
                            <div className="error-banner">
                                <AlertCircle size={20} />
                                <span>{errors.general}</span>
                            </div>
                        )}

                        {/* Username Field */}
                        <div className="form-group">
                            <label htmlFor="username" className="form-label">
                                <User size={18} />
                                Username
                            </label>
                            <input
                                id="username"
                                type="text"
                                className={`form-input ${errors.username ? 'error' : ''}`}
                                placeholder="Choose a username"
                                value={formData.username}
                                onChange={handleChange('username')}
                                disabled={loading}
                            />
                            {errors.username && (
                                <span className="error-message">{errors.username}</span>
                            )}
                        </div>

                        {/* Email Field */}
                        <div className="form-group">
                            <label htmlFor="email" className="form-label">
                                <Mail size={18} />
                                Email Address
                            </label>
                            <input
                                id="email"
                                type="email"
                                className={`form-input ${errors.email ? 'error' : ''}`}
                                placeholder="your.email@example.com"
                                value={formData.email}
                                onChange={handleChange('email')}
                                disabled={loading}
                            />
                            {errors.email && (
                                <span className="error-message">{errors.email}</span>
                            )}
                        </div>

                        {/* Password Field */}
                        <div className="form-group">
                            <label htmlFor="password" className="form-label">
                                <Lock size={18} />
                                Password
                            </label>
                            <input
                                id="password"
                                type="password"
                                className={`form-input ${errors.password ? 'error' : ''}`}
                                placeholder="Create a strong password"
                                value={formData.password}
                                onChange={handleChange('password')}
                                disabled={loading}
                            />
                            {errors.password && (
                                <span className="error-message">{errors.password}</span>
                            )}
                            <div className="password-requirements">
                                <span className={formData.password.length >= 8 ? 'met' : ''}>
                                    • At least 8 characters
                                </span>
                                <span className={/(?=.*[a-z])(?=.*[A-Z])/.test(formData.password) ? 'met' : ''}>
                                    • Upper & lowercase
                                </span>
                                <span className={/(?=.*\d)/.test(formData.password) ? 'met' : ''}>
                                    • At least 1 number
                                </span>
                            </div>
                        </div>

                        {/* Confirm Password Field */}
                        <div className="form-group">
                            <label htmlFor="confirmPassword" className="form-label">
                                <Lock size={18} />
                                Confirm Password
                            </label>
                            <input
                                id="confirmPassword"
                                type="password"
                                className={`form-input ${errors.confirmPassword ? 'error' : ''}`}
                                placeholder="Re-enter your password"
                                value={formData.confirmPassword}
                                onChange={handleChange('confirmPassword')}
                                disabled={loading}
                            />
                            {errors.confirmPassword && (
                                <span className="error-message">{errors.confirmPassword}</span>
                            )}
                        </div>

                        {/* Submit Button */}
                        <button
                            type="submit"
                            className="submit-button"
                            disabled={loading}
                        >
                            {loading ? (
                                <>
                                    <Loader2 size={20} className="spinner-icon" />
                                    Creating Account...
                                </>
                            ) : (
                                <>
                                    Create Account
                                    <ArrowRight size={20} />
                                </>
                            )}
                        </button>

                        {/* Sign In Link */}
                        <div className="form-footer">
                            <p>
                                Already have an account?{' '}
                                <button
                                    type="button"
                                    className="link-button"
                                    onClick={() => navigate('/login')}
                                    disabled={loading}
                                >
                                    Sign In
                                </button>
                            </p>
                        </div>
                    </form>
                </div>

                {/* Features Sidebar */}
                <div className="features-sidebar">
                    <h2 className="sidebar-title">Why Join Portfolio Insights?</h2>
                    <div className="feature-list">
                        <div className="feature-item">
                            <div className="feature-icon">
                                <CheckCircle2 size={24} />
                            </div>
                            <div className="feature-content">
                                <h3>Real-Time Tracking</h3>
                                <p>Monitor your investments with live market data</p>
                            </div>
                        </div>
                        <div className="feature-item">
                            <div className="feature-icon">
                                <CheckCircle2 size={24} />
                            </div>
                            <div className="feature-content">
                                <h3>Advanced Analytics</h3>
                                <p>Make informed decisions with powerful insights</p>
                            </div>
                        </div>
                        <div className="feature-item">
                            <div className="feature-icon">
                                <CheckCircle2 size={24} />
                            </div>
                            <div className="feature-content">
                                <h3>Secure & Private</h3>
                                <p>Bank-level security for your financial data</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default RegisterPage;
