import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import MobileMenu from '../MobileMenu';

describe('MobileMenu', () => {
    const mockNavLinks = [
        { label: 'Overview', href: '/', id: 'overview' as const },
        { label: 'Transactions', href: '/transactions', id: 'transactions' as const },
        { label: 'Asset Fundamentals', href: '/fundamentals', id: 'fundamentals' as const },
    ];

    const mockOnClose = vi.fn();

    beforeEach(() => {
        mockOnClose.mockClear();
    });

    afterEach(() => {
        // Reset body overflow
        document.body.style.overflow = 'unset';
    });

    it('renders nothing when isOpen is false', () => {
        const { container } = render(
            <MobileMenu
                isOpen={false}
                onClose={mockOnClose}
                navLinks={mockNavLinks}
                currentPage="overview"
            />
        );

        expect(container.firstChild).toBeNull();
    });

    it('renders menu when isOpen is true', () => {
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        expect(screen.getByRole('navigation', { name: 'Mobile navigation' })).toBeInTheDocument();
    });

    it('renders all navigation links', () => {
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        expect(screen.getByText('Overview')).toBeInTheDocument();
        expect(screen.getByText('Transactions')).toBeInTheDocument();
        expect(screen.getByText('Asset Fundamentals')).toBeInTheDocument();
    });

    it('renders backdrop', () => {
        const { container } = render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        // Backdrop is the first div with aria-hidden
        const backdrop = container.querySelector('[aria-hidden="true"]');
        expect(backdrop).toBeInTheDocument();
    });

    it('calls onClose when backdrop is clicked', async () => {
        const user = userEvent.setup();
        const { container } = render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        const backdrop = container.querySelector('[aria-hidden="true"]') as HTMLElement;
        await user.click(backdrop);

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('calls onClose when Escape key is pressed', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        await user.keyboard('{Escape}');

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('does not call onClose when Escape is pressed and menu is closed', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={false}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        await user.keyboard('{Escape}');

        expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('prevents body scroll when menu is open', () => {
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        expect(document.body.style.overflow).toBe('hidden');

        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={false}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        expect(document.body.style.overflow).toBe('unset');
    });

    it('highlights active page', () => {
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="transactions"
                />
            </MemoryRouter>
        );

        const transactionsLink = screen.getByText('Transactions');
        expect(transactionsLink).toHaveAttribute('aria-current', 'page');
    });

    it('calls onClose when a nav link is clicked', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        const link = screen.getByText('Transactions');
        await user.click(link);

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('displays version information', () => {
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        expect(screen.getByText('Portfolio Insights v1.0')).toBeInTheDocument();
    });

    it('has correct ARIA attributes', () => {
        render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        const nav = screen.getByRole('navigation', { name: 'Mobile navigation' });
        expect(nav).toHaveAttribute('id', 'mobile-menu');
    });

    it('cleans up event listeners on unmount', () => {
        const { unmount } = render(
            <MemoryRouter>
                <MobileMenu
                    isOpen={true}
                    onClose={mockOnClose}
                    navLinks={mockNavLinks}
                    currentPage="overview"
                />
            </MemoryRouter>
        );

        unmount();

        // Body overflow should be reset
        expect(document.body.style.overflow).toBe('unset');
    });
});
