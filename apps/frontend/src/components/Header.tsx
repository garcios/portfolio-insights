import { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { Wallet, Menu, X } from 'lucide-react';
import NavLink from './NavLink';
import MobileMenu from './MobileMenu';
import UserMenu from './UserMenu';

const Header = () => {
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
    const location = useLocation();

    // Determine current page ID based on path
    const getCurrentPageId = () => {
        const path = location.pathname;
        if (path === '/') return 'overview';
        if (path.startsWith('/transactions')) return 'transactions';
        if (path.startsWith('/fundamentals')) return 'fundamentals';
        return 'overview';
    };

    const currentPage = getCurrentPageId();

    const toggleMobileMenu = () => {
        setIsMobileMenuOpen(!isMobileMenuOpen);
    };

    const closeMobileMenu = () => {
        setIsMobileMenuOpen(false);
    };

    const navLinks = [
        { label: 'Overview', href: '/dashboard', id: 'overview' as const },
        { label: 'Transactions', href: '/transactions', id: 'transactions' as const },
        { label: 'Asset Fundamentals', href: '/fundamentals', id: 'fundamentals' as const },
    ];

    return (
        <header
            style={{
                background: 'var(--color-bg-secondary)',
                borderBottom: '1px solid var(--color-border)',
                position: 'sticky',
                top: 0,
                zIndex: 100,
                backdropFilter: 'blur(10px)',
            }}
        >
            <div
                style={{
                    maxWidth: '1400px',
                    margin: '0 auto',
                    padding: '16px 24px',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}
            >
                {/* Logo Section */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <div
                        style={{
                            width: '40px',
                            height: '40px',
                            borderRadius: '10px',
                            background: 'linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                        }}
                    >
                        <Wallet size={24} color="white" />
                    </div>
                    <div>
                        <h1
                            style={{
                                fontSize: '1.5rem',
                                fontWeight: '700',
                                color: 'var(--color-text-primary)',
                                lineHeight: '1',
                            }}
                        >
                            Portfolio Insights
                        </h1>
                        <p
                            style={{
                                fontSize: '0.75rem',
                                color: 'var(--color-text-tertiary)',
                                marginTop: '4px',
                            }}
                        >
                            Real-time portfolio tracking
                        </p>
                    </div>
                </div>

                {/* Desktop Navigation */}
                <nav
                    aria-label="Main navigation"
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                    }}
                    className="desktop-nav"
                >
                    {navLinks.map((link) => (
                        <NavLink
                            key={link.id}
                            href={link.href}
                            isActive={currentPage === link.id}
                        >
                            {link.label}
                        </NavLink>
                    ))}
                </nav>

                {/* Right Section - User Menu & Mobile Toggle */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    {/* User Menu (visible on all screens) */}
                    <UserMenu placement="bottom-end" />

                    {/* Mobile Menu Toggle */}
                    <button
                        onClick={toggleMobileMenu}
                        aria-label={isMobileMenuOpen ? 'Close menu' : 'Open menu'}
                        aria-expanded={isMobileMenuOpen}
                        aria-controls="mobile-menu"
                        className="mobile-menu-toggle"
                        style={{
                            padding: '10px',
                            borderRadius: '8px',
                            background: 'var(--color-bg-tertiary)',
                            border: '1px solid var(--color-border)',
                            color: 'var(--color-text-secondary)',
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                            display: 'none',
                        }}
                    >
                        {isMobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
                    </button>
                </div>
            </div>

            {/* Mobile Menu */}
            <MobileMenu
                isOpen={isMobileMenuOpen}
                onClose={closeMobileMenu}
                navLinks={navLinks}
                currentPage={currentPage}
            />
        </header>
    );
};

export default Header;
