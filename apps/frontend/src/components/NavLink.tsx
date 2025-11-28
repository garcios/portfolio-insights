import { ReactNode } from 'react';

interface NavLinkProps {
    href: string;
    isActive?: boolean;
    children: ReactNode;
    onClick?: () => void;
}

const NavLink = ({ href, isActive = false, children, onClick }: NavLinkProps) => {
    const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        e.preventDefault();
        if (onClick) {
            onClick();
        }
        // In a real app with routing, you'd navigate here
        // For now, we'll just prevent default
    };

    return (
        <a
            href={href}
            onClick={handleClick}
            aria-current={isActive ? 'page' : undefined}
            style={{
                padding: '10px 16px',
                borderRadius: '8px',
                fontSize: '0.875rem',
                fontWeight: isActive ? '600' : '500',
                color: isActive ? 'var(--color-text-primary)' : 'var(--color-text-secondary)',
                background: isActive ? 'var(--color-bg-tertiary)' : 'transparent',
                border: isActive ? '1px solid var(--color-border)' : '1px solid transparent',
                textDecoration: 'none',
                cursor: 'pointer',
                transition: 'all 0.2s',
                position: 'relative',
                display: 'inline-block',
            }}
            onMouseEnter={(e) => {
                if (!isActive) {
                    e.currentTarget.style.background = 'var(--color-bg-hover)';
                    e.currentTarget.style.borderColor = 'var(--color-border)';
                    e.currentTarget.style.color = 'var(--color-text-primary)';
                }
            }}
            onMouseLeave={(e) => {
                if (!isActive) {
                    e.currentTarget.style.background = 'transparent';
                    e.currentTarget.style.borderColor = 'transparent';
                    e.currentTarget.style.color = 'var(--color-text-secondary)';
                }
            }}
        >
            {children}
            {isActive && (
                <span
                    style={{
                        position: 'absolute',
                        bottom: '-1px',
                        left: '50%',
                        transform: 'translateX(-50%)',
                        width: '60%',
                        height: '2px',
                        background: 'linear-gradient(90deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                        borderRadius: '2px',
                    }}
                />
            )}
        </a>
    );
};

export default NavLink;
