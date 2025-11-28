import { useEffect } from 'react';
import NavLink from './NavLink';

interface MobileMenuProps {
    isOpen: boolean;
    onClose: () => void;
    navLinks: Array<{
        label: string;
        href: string;
        id: 'overview' | 'transactions' | 'fundamentals';
    }>;
    currentPage: 'overview' | 'transactions' | 'fundamentals';
}

const MobileMenu = ({ isOpen, onClose, navLinks, currentPage }: MobileMenuProps) => {
    // Close menu on escape key
    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape' && isOpen) {
                onClose();
            }
        };

        document.addEventListener('keydown', handleEscape);
        return () => document.removeEventListener('keydown', handleEscape);
    }, [isOpen, onClose]);

    // Prevent body scroll when menu is open
    useEffect(() => {
        if (isOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = 'unset';
        }

        return () => {
            document.body.style.overflow = 'unset';
        };
    }, [isOpen]);

    if (!isOpen) return null;

    return (
        <>
            {/* Backdrop */}
            <div
                onClick={onClose}
                style={{
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    background: 'rgba(0, 0, 0, 0.5)',
                    backdropFilter: 'blur(4px)',
                    zIndex: 999,
                    animation: 'fadeIn 0.2s ease-out',
                }}
                aria-hidden="true"
            />

            {/* Mobile Menu Panel */}
            <nav
                id="mobile-menu"
                role="navigation"
                aria-label="Mobile navigation"
                style={{
                    position: 'fixed',
                    top: '73px', // Height of header
                    right: 0,
                    bottom: 0,
                    width: '280px',
                    maxWidth: '80vw',
                    background: 'var(--color-bg-card)',
                    borderLeft: '1px solid var(--color-border)',
                    boxShadow: 'var(--shadow-xl)',
                    zIndex: 1000,
                    padding: '24px',
                    overflowY: 'auto',
                    animation: 'slideInRight 0.3s ease-out',
                }}
            >
                <div
                    style={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '8px',
                    }}
                >
                    {navLinks.map((link) => (
                        <div key={link.id} style={{ width: '100%' }}>
                            <NavLink
                                href={link.href}
                                isActive={currentPage === link.id}
                                onClick={onClose}
                            >
                                {link.label}
                            </NavLink>
                        </div>
                    ))}
                </div>

                {/* Optional: Add additional mobile menu content */}
                <div
                    style={{
                        marginTop: '32px',
                        paddingTop: '24px',
                        borderTop: '1px solid var(--color-border)',
                    }}
                >
                    <p
                        style={{
                            fontSize: '0.75rem',
                            color: 'var(--color-text-tertiary)',
                            textAlign: 'center',
                        }}
                    >
                        Portfolio Insights v1.0
                    </p>
                </div>
            </nav>

            <style>{`
                @keyframes slideInRight {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
            `}</style>
        </>
    );
};

export default MobileMenu;
