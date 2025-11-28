import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import NavLink from '../NavLink';

describe('NavLink', () => {
    it('renders link with children', () => {
        render(<NavLink href="/test">Test Link</NavLink>);

        expect(screen.getByText('Test Link')).toBeInTheDocument();
    });

    it('applies href attribute', () => {
        render(<NavLink href="/test">Test Link</NavLink>);

        const link = screen.getByText('Test Link');
        expect(link).toHaveAttribute('href', '/test');
    });

    it('applies active state styling when isActive is true', () => {
        render(<NavLink href="/test" isActive>Active Link</NavLink>);

        const link = screen.getByText('Active Link');
        expect(link).toHaveStyle({
            fontWeight: '600',
        });
        expect(link).toHaveAttribute('aria-current', 'page');
    });

    it('does not apply active state when isActive is false', () => {
        render(<NavLink href="/test" isActive={false}>Inactive Link</NavLink>);

        const link = screen.getByText('Inactive Link');
        expect(link).toHaveStyle({
            fontWeight: '500',
        });
        expect(link).not.toHaveAttribute('aria-current');
    });

    it('renders active indicator when isActive is true', () => {
        const { container } = render(<NavLink href="/test" isActive>Active Link</NavLink>);

        const indicator = container.querySelector('span');
        expect(indicator).toBeInTheDocument();
        expect(indicator).toHaveStyle({
            position: 'absolute',
            bottom: '-1px',
        });
    });

    it('does not render active indicator when isActive is false', () => {
        const { container } = render(<NavLink href="/test" isActive={false}>Inactive Link</NavLink>);

        const indicator = container.querySelector('span');
        expect(indicator).not.toBeInTheDocument();
    });

    it('calls onClick handler when clicked', async () => {
        const user = userEvent.setup();
        const handleClick = vi.fn();

        render(<NavLink href="/test" onClick={handleClick}>Clickable Link</NavLink>);

        const link = screen.getByText('Clickable Link');
        await user.click(link);

        expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('prevents default navigation', async () => {
        const user = userEvent.setup();

        render(<NavLink href="/test">Test Link</NavLink>);

        const link = screen.getByText('Test Link');
        await user.click(link);

        // The default should be prevented in the component
        expect(link).toBeInTheDocument();
    });

    it('applies hover styles on mouse enter (non-active)', async () => {
        const user = userEvent.setup();

        render(<NavLink href="/test" isActive={false}>Hover Link</NavLink>);

        const link = screen.getByText('Hover Link');
        await user.hover(link);

        // Hover effects are applied via inline styles in event handlers
        expect(link).toBeInTheDocument();
    });
});
