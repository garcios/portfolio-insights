import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Header from '../Header';

describe('Header', () => {
    beforeEach(() => {
        // Reset body overflow before each test
        document.body.style.overflow = 'unset';
    });

    it('renders the app logo and title', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        expect(screen.getByText('Portfolio Insights')).toBeInTheDocument();
        expect(screen.getByText('Real-time portfolio tracking')).toBeInTheDocument();
    });

    it('renders all navigation links', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        expect(screen.getByText('Overview')).toBeInTheDocument();
        expect(screen.getByText('Transactions')).toBeInTheDocument();
        expect(screen.getByText('Asset Fundamentals')).toBeInTheDocument();
    });

    it('highlights the current page', () => {
        render(
            <MemoryRouter initialEntries={['/transactions']}>
                <Header />
            </MemoryRouter>
        );

        const transactionsLink = screen.getByText('Transactions');
        expect(transactionsLink).toHaveAttribute('aria-current', 'page');
    });

    it('defaults to overview page when currentPage is not provided', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const overviewLink = screen.getByText('Overview');
        expect(overviewLink).toHaveAttribute('aria-current', 'page');
    });

    it('renders user menu', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const userMenuButton = screen.getByRole('button', { name: 'User menu' });
        expect(userMenuButton).toBeInTheDocument();
    });

    it('renders mobile menu toggle button', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const mobileMenuButton = screen.getByLabelText('Open menu');
        expect(mobileMenuButton).toBeInTheDocument();
    });

    it('mobile menu is closed by default', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        // Mobile menu should not be visible initially
        expect(screen.queryByRole('navigation', { name: 'Mobile navigation' })).not.toBeInTheDocument();
    });

    it('opens mobile menu when toggle button is clicked', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const mobileMenuButton = screen.getByLabelText('Open menu');
        await user.click(mobileMenuButton);

        expect(screen.getByRole('navigation', { name: 'Mobile navigation' })).toBeInTheDocument();
    });

    it('closes mobile menu when toggle button is clicked again', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const openButton = screen.getByLabelText('Open menu');
        await user.click(openButton);

        expect(screen.getByRole('navigation', { name: 'Mobile navigation' })).toBeInTheDocument();

        const closeButton = screen.getByLabelText('Close menu');
        await user.click(closeButton);

        expect(screen.queryByRole('navigation', { name: 'Mobile navigation' })).not.toBeInTheDocument();
    });

    it('updates mobile menu button aria-expanded attribute', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const mobileMenuButton = screen.getByLabelText('Open menu');
        expect(mobileMenuButton).toHaveAttribute('aria-expanded', 'false');

        await user.click(mobileMenuButton);

        const closeButton = screen.getByLabelText('Close menu');
        expect(closeButton).toHaveAttribute('aria-expanded', 'true');
    });

    it('has correct aria-controls attribute on mobile menu button', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const mobileMenuButton = screen.getByLabelText('Open menu');
        expect(mobileMenuButton).toHaveAttribute('aria-controls', 'mobile-menu');
    });

    it('renders main navigation with correct aria-label', () => {
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const mainNav = screen.getByRole('navigation', { name: 'Main navigation' });
        expect(mainNav).toBeInTheDocument();
    });

    it('passes correct props to MobileMenu', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter initialEntries={['/fundamentals']}>
                <Header />
            </MemoryRouter>
        );

        const mobileMenuButton = screen.getByLabelText('Open menu');
        await user.click(mobileMenuButton);

        // Check that the mobile menu shows the correct active page
        const fundamentalsLinks = screen.getAllByText('Asset Fundamentals');
        const fundamentalsLink = fundamentalsLinks[fundamentalsLinks.length - 1]; // Last one is in mobile menu

        expect(fundamentalsLink).toHaveAttribute('aria-current', 'page');
    });

    it('closes mobile menu when a nav link is clicked', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        // Open mobile menu
        const openButton = screen.getByLabelText('Open menu');
        await user.click(openButton);

        expect(screen.getByRole('navigation', { name: 'Mobile navigation' })).toBeInTheDocument();

        // Click a link in the mobile menu - get all Transactions links and click the one in mobile menu
        const allTransactionsLinks = screen.getAllByText('Transactions');
        const transactionsLink = allTransactionsLinks[allTransactionsLinks.length - 1]; // Last one is in mobile menu
        await user.click(transactionsLink);

        // Mobile menu should be closed
        expect(screen.queryByRole('navigation', { name: 'Mobile navigation' })).not.toBeInTheDocument();
    });

    it('renders logo icon', () => {
        const { container } = render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        // Check for Wallet icon (SVG)
        const svgs = container.querySelectorAll('svg');
        expect(svgs.length).toBeGreaterThan(0);
    });

    it('has sticky positioning', () => {
        const { container } = render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const header = container.querySelector('header');
        expect(header).toHaveStyle({
            position: 'sticky',
            top: '0',
        });
    });

    it('applies correct z-index for stacking', () => {
        const { container } = render(
            <MemoryRouter>
                <Header />
            </MemoryRouter>
        );

        const header = container.querySelector('header');
        expect(header).toHaveStyle({
            zIndex: '100',
        });
    });
});
