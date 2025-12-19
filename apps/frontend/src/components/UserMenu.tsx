import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { User, Settings, Bell, LogOut, Moon, Sun } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';

const UserMenu = () => {
    const navigate = useNavigate();
    const [isOpen, setIsOpen] = useState(false);
    const [isDarkMode, setIsDarkMode] = useState(true); // Default to dark mode
    const menuRef = useRef<HTMLDivElement>(null);
    const { user, logout } = useAuth();

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [isOpen]);

    // Close menu on escape key
    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape' && isOpen) {
                setIsOpen(false);
            }
        };

        document.addEventListener('keydown', handleEscape);
        return () => document.removeEventListener('keydown', handleEscape);
    }, [isOpen]);

    const toggleMenu = () => {
        setIsOpen(!isOpen);
    };

    const toggleTheme = () => {
        setIsDarkMode(!isDarkMode);
        // In a real app, you'd update the theme context/state here
    };

    const handleLogout = () => {
        logout();
        setIsOpen(false);
    };

    const menuItems = [
        { icon: Settings, label: 'Settings', onClick: () => navigate('/settings') },
        { icon: Bell, label: 'Notifications', onClick: () => console.log('Notifications') },
        { icon: isDarkMode ? Sun : Moon, label: isDarkMode ? 'Light Mode' : 'Dark Mode', onClick: toggleTheme },
        { icon: LogOut, label: 'Sign Out', onClick: handleLogout, isDanger: true },
    ];

    return (
        <div ref={menuRef} style={{ position: 'relative' }}>
            {/* User Avatar Button */}
            <button
                onClick={toggleMenu}
                aria-label="User menu"
                aria-expanded={isOpen}
                aria-haspopup="true"
                style={{
                    width: '36px',
                    height: '36px',
                    borderRadius: '50%',
                    background: 'linear-gradient(135deg, var(--color-accent) 0%, var(--color-primary) 100%)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer',
                    border: '2px solid var(--color-border)',
                    transition: 'all 0.2s',
                    outline: 'none',
                }}
                onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'scale(1.05)';
                    e.currentTarget.style.boxShadow = '0 0 20px rgba(99, 102, 241, 0.4)';
                }}
                onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'scale(1)';
                    e.currentTarget.style.boxShadow = 'none';
                }}
            >
                <User size={18} color="white" />
            </button>

            {/* Dropdown Menu */}
            {isOpen && (
                <div
                    role="menu"
                    style={{
                        position: 'absolute',
                        top: 'calc(100% + 8px)',
                        right: 0,
                        width: '220px',
                        background: 'var(--color-bg-card)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--radius-lg)',
                        boxShadow: 'var(--shadow-xl)',
                        padding: '8px',
                        zIndex: 1000,
                        animation: 'fadeInDown 0.2s ease-out',
                    }}
                >
                    {/* User Info */}
                    <div
                        style={{
                            padding: '12px',
                            borderBottom: '1px solid var(--color-border)',
                            marginBottom: '8px',
                        }}
                    >
                        <p
                            style={{
                                fontSize: '0.875rem',
                                fontWeight: '600',
                                color: 'var(--color-text-primary)',
                                marginBottom: '4px',
                            }}
                        >
                            {user?.firstName ? `${user.firstName} ${user.lastName || ''}` : (user?.username || user?.email?.split('@')[0] || 'User')}
                        </p>
                        <p
                            style={{
                                fontSize: '0.75rem',
                                color: 'var(--color-text-tertiary)',
                            }}
                        >
                            {user?.role ? `${user.role} • ` : ''}{user?.email || 'user@portfolio.com'}
                        </p>
                    </div>

                    {/* Menu Items */}
                    {menuItems.map((item, index) => {
                        const Icon = item.icon;
                        return (
                            <button
                                key={index}
                                role="menuitem"
                                onClick={() => {
                                    item.onClick();
                                    if (item.label !== 'Light Mode' && item.label !== 'Dark Mode') {
                                        setIsOpen(false);
                                    }
                                }}
                                style={{
                                    width: '100%',
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '12px',
                                    padding: '10px 12px',
                                    borderRadius: 'var(--radius-md)',
                                    background: 'transparent',
                                    border: 'none',
                                    color: item.isDanger ? 'var(--color-danger)' : 'var(--color-text-secondary)',
                                    fontSize: '0.875rem',
                                    fontWeight: '500',
                                    cursor: 'pointer',
                                    transition: 'all 0.2s',
                                    textAlign: 'left',
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.background = 'var(--color-bg-hover)';
                                    e.currentTarget.style.color = item.isDanger
                                        ? 'var(--color-danger)'
                                        : 'var(--color-text-primary)';
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.background = 'transparent';
                                    e.currentTarget.style.color = item.isDanger
                                        ? 'var(--color-danger)'
                                        : 'var(--color-text-secondary)';
                                }}
                            >
                                <Icon size={16} />
                                <span>{item.label}</span>
                            </button>
                        );
                    })}
                </div>
            )}

            <style>{`
                @keyframes fadeInDown {
                    from {
                        opacity: 0;
                        transform: translateY(-10px);
                    }
                    to {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }
            `}</style>
        </div>
    );
};

export default UserMenu;
