import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import UserMenu from '../UserMenu';

describe('UserMenu', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        // Mock console.log to avoid cluttering test output
        vi.spyOn(console, 'log').mockImplementation(() => { });
    });

    it('renders user avatar button', () => {
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        expect(button).toBeInTheDocument();
    });

    it('menu is closed by default', () => {
        render(<UserMenu />);

        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('opens menu when avatar is clicked', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        expect(screen.getByRole('menu')).toBeInTheDocument();
    });

    it('closes menu when avatar is clicked again', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });

        // Open menu
        await user.click(button);
        expect(screen.getByRole('menu')).toBeInTheDocument();

        // Close menu
        await user.click(button);
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('displays user information when menu is open', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        expect(screen.getByText('Demo User')).toBeInTheDocument();
        expect(screen.getByText('demo@portfolio.com')).toBeInTheDocument();
    });

    it('displays all menu items', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        expect(screen.getByText('Settings')).toBeInTheDocument();
        expect(screen.getByText('Notifications')).toBeInTheDocument();
        expect(screen.getByText('Light Mode')).toBeInTheDocument(); // Default is dark mode
        expect(screen.getByText('Sign Out')).toBeInTheDocument();
    });

    it('calls Settings handler when Settings is clicked', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const settingsButton = screen.getByText('Settings');
        await user.click(settingsButton);

        expect(console.log).toHaveBeenCalledWith('Settings');
    });

    it('calls Notifications handler when Notifications is clicked', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const notificationsButton = screen.getByText('Notifications');
        await user.click(notificationsButton);

        expect(console.log).toHaveBeenCalledWith('Notifications');
    });

    it('calls Sign Out handler when Sign Out is clicked', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const signOutButton = screen.getByText('Sign Out');
        await user.click(signOutButton);

        expect(console.log).toHaveBeenCalledWith('Sign out');
    });

    it('toggles theme when theme button is clicked', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        // Initially shows "Light Mode" (dark mode is active)
        expect(screen.getByText('Light Mode')).toBeInTheDocument();

        const themeButton = screen.getByText('Light Mode');
        await user.click(themeButton);

        // After toggle, should show "Dark Mode" (light mode is active)
        expect(screen.getByText('Dark Mode')).toBeInTheDocument();
    });

    it('keeps menu open when theme is toggled', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const themeButton = screen.getByText('Light Mode');
        await user.click(themeButton);

        // Menu should still be open
        expect(screen.getByRole('menu')).toBeInTheDocument();
    });

    it('closes menu when non-theme menu item is clicked', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const settingsButton = screen.getByText('Settings');
        await user.click(settingsButton);

        // Menu should be closed
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('closes menu when clicking outside', async () => {
        const user = userEvent.setup();
        render(
            <div>
                <UserMenu />
                <div data-testid="outside">Outside element</div>
            </div>
        );

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        expect(screen.getByRole('menu')).toBeInTheDocument();

        const outsideElement = screen.getByTestId('outside');
        await user.click(outsideElement);

        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('closes menu when Escape key is pressed', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        expect(screen.getByRole('menu')).toBeInTheDocument();

        await user.keyboard('{Escape}');

        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('does not close menu when Escape is pressed and menu is already closed', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        expect(screen.queryByRole('menu')).not.toBeInTheDocument();

        await user.keyboard('{Escape}');

        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('has correct ARIA attributes on button', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });

        expect(button).toHaveAttribute('aria-haspopup', 'true');
        expect(button).toHaveAttribute('aria-expanded', 'false');

        await user.click(button);

        expect(button).toHaveAttribute('aria-expanded', 'true');
    });

    it('renders menu items with correct roles', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const menu = screen.getByRole('menu');
        const menuItems = within(menu).getAllByRole('menuitem');

        expect(menuItems).toHaveLength(4); // Settings, Notifications, Theme, Sign Out
    });

    it('renders icons for all menu items', async () => {
        const user = userEvent.setup();
        render(<UserMenu />);

        const button = screen.getByRole('button', { name: 'User menu' });
        await user.click(button);

        const menu = screen.getByRole('menu');
        const icons = menu.querySelectorAll('svg');

        // Should have icons for: Settings, Notifications, Theme, Sign Out, plus User icon in button
        expect(icons.length).toBeGreaterThanOrEqual(4);
    });
});
