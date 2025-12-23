import { useState } from 'react';
import { useMutation, useQuery } from '@apollo/client';
import { useAuth } from '../auth/AuthContext';
import { UPDATE_USER } from '../graphql/mutations';
import { GET_USER } from '../graphql/queries';
import { User, Mail, Lock, Save, AlertCircle, CheckCircle, Globe } from 'lucide-react';



const SettingsPage = () => {
    const { user } = useAuth();
    const [formData, setFormData] = useState({
        firstName: user?.firstName || '',
        lastName: user?.lastName || '',
        username: user?.username || '',
        email: user?.email || '',
        defaultCurrency: 'AUD',
        dateFormat: 'DD/MM/YYYY',
    });
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const [updateUser, { loading }] = useMutation(UPDATE_USER, {
        onCompleted: () => {
            setMessage({ type: 'success', text: 'Profile updated successfully' });
            // Clear message after 3 seconds
            setTimeout(() => setMessage(null), 3000);
        },
        onError: (error) => {
            setMessage({ type: 'error', text: error.message || 'Failed to update profile' });
        },
    });

    useQuery(GET_USER, {
        variables: { id: user?.id },
        skip: !user?.id,
        fetchPolicy: 'network-only', // Ensure we get the latest data
        onCompleted: (data) => {
            if (data?.user) {
                const u = data.user;
                const prefs = u.preferences || {};
                setFormData({
                    firstName: u.firstName || '',
                    lastName: u.lastName || '',
                    username: u.username || '',
                    email: u.email || '',
                    defaultCurrency: prefs.default_currency || 'AUD',
                    dateFormat: prefs.date_format || 'DD/MM/YYYY',
                });
            }
        },
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setMessage(null);

        try {
            await updateUser({
                variables: {
                    input: {
                        firstName: formData.firstName,
                        lastName: formData.lastName,
                        username: formData.username,
                        email: formData.email,
                        preferences: {
                            default_currency: formData.defaultCurrency,
                            date_format: formData.dateFormat,
                        }
                    }
                }
            });
        } catch (err) {
            console.error('Error updating user:', err);
        }
    };

    return (
        <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
            <div style={{ maxWidth: '800px', margin: '40px auto', padding: '0 20px' }}>
                <div style={{ marginBottom: '32px' }}>
                    <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: 'var(--color-text-primary)' }}>Settings</h1>
                    <p style={{ color: 'var(--color-text-secondary)', marginTop: '8px' }}>Manage your account settings and preferences.</p>
                </div>

                <div style={{
                    background: 'var(--color-bg-secondary)',
                    borderRadius: '12px',
                    border: '1px solid var(--color-border)',
                    padding: '24px',
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '24px', paddingBottom: '16px', borderBottom: '1px solid var(--color-border)' }}>
                        <div style={{
                            width: '40px',
                            height: '40px',
                            borderRadius: '50%',
                            background: 'var(--color-bg-tertiary)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: 'var(--color-primary)'
                        }}>
                            <User size={20} />
                        </div>
                        <div>
                            <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: 'var(--color-text-primary)' }}>Profile Information</h2>
                            <p style={{ fontSize: '0.875rem', color: 'var(--color-text-secondary)' }}>Update your personal details.</p>
                        </div>
                    </div>

                    {message && (
                        <div style={{
                            padding: '12px 16px',
                            borderRadius: '8px',
                            marginBottom: '24px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            background: message.type === 'success' ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)',
                            color: message.type === 'success' ? '#10B981' : '#EF4444',
                            border: `1px solid ${message.type === 'success' ? 'rgba(16, 185, 129, 0.2)' : 'rgba(239, 68, 68, 0.2)'}`
                        }}>
                            {message.type === 'success' ? <CheckCircle size={18} /> : <AlertCircle size={18} />}
                            <span style={{ fontSize: '0.875rem', fontWeight: '500' }}>{message.text}</span>
                        </div>
                    )}

                    <form onSubmit={handleSubmit}>
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '20px', marginBottom: '24px' }}>
                            <div>
                                <label htmlFor="firstName" style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                                    First Name
                                </label>
                                <div style={{ position: 'relative' }}>
                                    <input
                                        type="text"
                                        id="firstName"
                                        name="firstName"
                                        value={formData.firstName}
                                        onChange={handleChange}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px 10px 36px',
                                            borderRadius: '8px',
                                            border: '1px solid var(--color-border)',
                                            background: 'var(--color-bg-tertiary)',
                                            color: 'var(--color-text-primary)',
                                            fontSize: '0.875rem'
                                        }}
                                    />
                                    <User size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                                </div>
                            </div>

                            <div>
                                <label htmlFor="lastName" style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                                    Last Name
                                </label>
                                <div style={{ position: 'relative' }}>
                                    <input
                                        type="text"
                                        id="lastName"
                                        name="lastName"
                                        value={formData.lastName}
                                        onChange={handleChange}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px 10px 36px',
                                            borderRadius: '8px',
                                            border: '1px solid var(--color-border)',
                                            background: 'var(--color-bg-tertiary)',
                                            color: 'var(--color-text-primary)',
                                            fontSize: '0.875rem'
                                        }}
                                    />
                                    <User size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                                </div>
                            </div>

                            <div>
                                <label htmlFor="username" style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                                    Username
                                </label>
                                <div style={{ position: 'relative' }}>
                                    <input
                                        type="text"
                                        id="username"
                                        name="username"
                                        value={formData.username}
                                        onChange={handleChange}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px 10px 36px',
                                            borderRadius: '8px',
                                            border: '1px solid var(--color-border)',
                                            background: 'var(--color-bg-tertiary)',
                                            color: 'var(--color-text-primary)',
                                            fontSize: '0.875rem'
                                        }}
                                    />
                                    <Lock size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                                </div>
                            </div>

                            <div>
                                <label htmlFor="email" style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                                    Email Address
                                </label>
                                <div style={{ position: 'relative' }}>
                                    <input
                                        type="email"
                                        id="email"
                                        name="email"
                                        value={formData.email}
                                        onChange={handleChange}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px 10px 36px',
                                            borderRadius: '8px',
                                            border: '1px solid var(--color-border)',
                                            background: 'var(--color-bg-tertiary)',
                                            color: 'var(--color-text-primary)',
                                            fontSize: '0.875rem'
                                        }}
                                    />
                                    <Mail size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                                </div>
                            </div>
                        </div>

                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '24px', paddingBottom: '16px', borderBottom: '1px solid var(--color-border)' }}>
                            <div style={{
                                width: '40px',
                                height: '40px',
                                borderRadius: '50%',
                                background: 'var(--color-bg-tertiary)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                color: 'var(--color-primary)'
                            }}>
                                <Globe size={20} />
                            </div>
                            <div>
                                <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: 'var(--color-text-primary)' }}>Preferences</h2>
                                <p style={{ fontSize: '0.875rem', color: 'var(--color-text-secondary)' }}>Customize your application experience.</p>
                            </div>
                        </div>

                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '20px', marginBottom: '24px' }}>
                            <div>
                                <label htmlFor="defaultCurrency" style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                                    Default Currency
                                </label>
                                <div style={{ position: 'relative' }}>
                                    <select
                                        id="defaultCurrency"
                                        name="defaultCurrency"
                                        value={formData.defaultCurrency}
                                        onChange={handleChange}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px 10px 36px',
                                            borderRadius: '8px',
                                            border: '1px solid var(--color-border)',
                                            background: 'var(--color-bg-tertiary)',
                                            color: 'var(--color-text-primary)',
                                            fontSize: '0.875rem',
                                            appearance: 'none',
                                            cursor: 'pointer'
                                        }}
                                    >
                                        <option value="USD">USD - US Dollar</option>
                                        <option value="EUR">EUR - Euro</option>
                                        <option value="GBP">GBP - British Pound</option>
                                        <option value="AUD">AUD - Australian Dollar</option>
                                        <option value="CAD">CAD - Canadian Dollar</option>
                                        <option value="PHP">PHP - Philippine Peso</option>
                                    </select>
                                    <Globe size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                                </div>
                            </div>

                            <div>
                                <label htmlFor="dateFormat" style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                                    Date Format
                                </label>
                                <div style={{ position: 'relative' }}>
                                    <select
                                        id="dateFormat"
                                        name="dateFormat"
                                        value={formData.dateFormat}
                                        onChange={handleChange}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px 10px 36px',
                                            borderRadius: '8px',
                                            border: '1px solid var(--color-border)',
                                            background: 'var(--color-bg-tertiary)',
                                            color: 'var(--color-text-primary)',
                                            fontSize: '0.875rem',
                                            appearance: 'none',
                                            cursor: 'pointer'
                                        }}
                                    >
                                        <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                                        <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                                        <option value="YYYY-MM-DD">YYYY-MM-DD</option>
                                    </select>
                                    <Globe size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                                </div>
                            </div>
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                            <button
                                type="submit"
                                disabled={loading}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    padding: '10px 20px',
                                    background: 'var(--color-primary)',
                                    color: 'white',
                                    border: 'none',
                                    borderRadius: '8px',
                                    fontSize: '0.875rem',
                                    fontWeight: '600',
                                    cursor: loading ? 'not-allowed' : 'pointer',
                                    opacity: loading ? 0.7 : 1,
                                    transition: 'background 0.2s',
                                }}
                            >
                                {loading ? (
                                    <>Saving...</>
                                ) : (
                                    <>
                                        <Save size={18} />
                                        Save Changes
                                    </>
                                )}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
};

export default SettingsPage;
