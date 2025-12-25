import { useState, useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../components/Sidebar';
import { Menu } from 'lucide-react';

const MainLayout = () => {
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);

    // Handle desktop view to ensure sidebar is visible
    useEffect(() => {
        const handleResize = () => {
            // We'll handle the responsive show/hide logic via CSS classes mostly,
            // but this state is useful for the mobile toggle
            if (window.innerWidth >= 768) {
                setIsSidebarOpen(true);
            } else {
                setIsSidebarOpen(false);
            }
        };

        // Initial check
        handleResize();

        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    return (
        <div style={{
            display: 'flex',
            minHeight: '100vh',
            background: 'var(--color-bg-primary)'
        }}>
            {/* Sidebar */}
            <Sidebar isOpen={isSidebarOpen} setIsOpen={setIsSidebarOpen} />

            {/* Main Content Area */}
            <div style={{
                flex: 1,
                marginLeft: window.innerWidth >= 768 ? '250px' : '0', // Adjust margin based on screen size (this is better done with CSS media queries)
                transition: 'margin-left 0.3s ease-in-out',
                display: 'flex',
                flexDirection: 'column',
                flexDirection: 'column',
            }}
                className="main-content"
            >
                {/* Mobile Header / Toggle */}
                <div
                    className="md:hidden"
                    style={{
                        display: window.innerWidth < 768 ? 'flex' : 'none', // Simple JS check, ideally CSS
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '16px 24px',
                        background: 'var(--color-bg-secondary)',
                        borderBottom: '1px solid var(--color-border)',
                        position: 'sticky',
                        top: 0,
                        zIndex: 30,
                    }}
                >
                    <button
                        onClick={() => setIsSidebarOpen(true)}
                        style={{
                            background: 'none',
                            border: 'none',
                            color: 'var(--color-text-primary)',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                        }}
                    >
                        <Menu size={24} />
                    </button>
                    <span style={{ fontWeight: 'bold', color: 'var(--color-text-primary)' }}>Portfolio Insights</span>
                    <div style={{ width: '24px' }}></div> {/* Spacer for centering */}
                </div>

                {/* Page Content */}
                <main style={{ flex: 1 }}>
                    <Outlet />
                </main>
            </div>

            {/* Inject a style tag for media queries since we are using inline styles mostly, 
                but we need media queries for the responsive layout behavior 
            */}
            <style>{`
                @media (min-width: 768px) {
                    aside {
                        transform: translateX(0) !important;
                    }
                    .main-content {
                        margin-left: 250px !important;
                    }
                    .md\\:hidden {
                        display: none !important;
                    }
                }
            `}</style>
        </div>
    );
};

export default MainLayout;
