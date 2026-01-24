import { useLocation } from 'react-router-dom';
import { Wallet, LayoutDashboard, ArrowRightLeft, BookOpen, Settings } from 'lucide-react';
import NavLink from './NavLink';
import UserMenu from './UserMenu';

// Define the props for the Sidebar
interface SidebarProps {
    className?: string;
    isOpen: boolean;
    setIsOpen: (isOpen: boolean) => void;
}

const Sidebar = ({ isOpen, setIsOpen }: SidebarProps) => {
    const location = useLocation();

    // Determine current page ID based on path
    const getCurrentPageId = () => {
        const path = location.pathname;
        if (path === '/') return 'overview';
        if (path.startsWith('/dashboard')) return 'overview';
        if (path.startsWith('/transactions')) return 'transactions';
        if (path.startsWith('/fundamentals')) return 'fundamentals';
        if (path.startsWith('/settings')) return 'settings';
        return 'overview';
    };

    const currentPage = getCurrentPageId();

    const navLinks = [
        { label: 'Overview', href: '/dashboard', id: 'overview' as const, icon: <LayoutDashboard size={20} /> },
        { label: 'Transactions', href: '/transactions', id: 'transactions' as const, icon: <ArrowRightLeft size={20} /> },
        { label: 'Asset Fundamentals', href: '/fundamentals', id: 'fundamentals' as const, icon: <BookOpen size={20} /> },
        { label: 'Settings', href: '/settings', id: 'settings' as const, icon: <Settings size={20} /> },
    ];

    return (
        <>
            {/* Mobile Overlay */}
            {isOpen && (
                <div
                    className="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden"
                    onClick={() => setIsOpen(false)}
                    style={{
                        position: 'fixed',
                        inset: 0,
                        backgroundColor: 'rgba(0, 0, 0, 0.5)',
                        zIndex: 40,
                    }}
                />
            )}

            {/* Sidebar Container */}
            <aside
                style={{
                    width: '250px',
                    height: '100vh',
                    background: 'var(--color-bg-secondary)',
                    borderRight: '1px solid var(--color-border)',
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    zIndex: 50,
                    display: 'flex',
                    flexDirection: 'column',
                    transform: isOpen ? 'translateX(0)' : 'translateX(-100%)',
                    transition: 'transform 0.3s ease-in-out',
                }}
                className={`sidebar ${isOpen ? 'open' : ''} md:translate-x-0`}
            >
                {/* Logo Section */}
                <div style={{
                    padding: '24px',
                    borderBottom: '1px solid var(--color-border)',
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '8px' }}>
                        <div
                            style={{
                                width: '32px',
                                height: '32px',
                                borderRadius: '8px',
                                background: 'linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                flexShrink: 0,
                            }}
                        >
                            <Wallet size={18} color="white" />
                        </div>
                        <h1
                            style={{
                                fontSize: '1.25rem',
                                fontWeight: '700',
                                color: 'var(--color-text-primary)',
                                lineHeight: '1',
                                whiteSpace: 'nowrap',
                            }}
                        >
                            Portfolio
                        </h1>
                    </div>
                </div>

                {/* Navigation Links */}
                <nav
                    style={{
                        flex: 1,
                        padding: '24px 16px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '8px',
                        overflowY: 'auto',
                    }}
                >
                    {navLinks.map((link) => (
                        <div key={link.id} style={{ display: 'flex' }}>
                            {/* We might need to adjust NavLink to support icons or just use a custom link here */}
                            <NavLink
                                href={link.href}
                                isActive={currentPage === link.id}
                            >
                                <span style={{ display: 'flex', alignItems: 'center', gap: '12px', width: '100%' }}>
                                    {link.icon || null}
                                    {link.label}
                                </span>
                            </NavLink>
                        </div>
                    ))}
                </nav>

                {/* User Section at Bottom */}
                <div style={{
                    padding: '16px',
                    borderTop: '1px solid var(--color-border)',
                }}>
                    <UserMenu placement="top-end" />
                </div>
            </aside>
        </>
    );
};

export default Sidebar;
