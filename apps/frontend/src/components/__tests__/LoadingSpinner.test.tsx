import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LoadingSpinner from '../LoadingSpinner';

describe('LoadingSpinner', () => {
    it('renders loading spinner', () => {
        render(<LoadingSpinner />);

        const spinner = document.querySelector('.spinner');
        expect(spinner).toBeInTheDocument();
    });

    it('displays loading message', () => {
        render(<LoadingSpinner />);

        expect(screen.getByText('Loading portfolio data...')).toBeInTheDocument();
    });

    it('applies correct styling', () => {
        const { container } = render(<LoadingSpinner />);

        const wrapper = container.firstChild as HTMLElement;
        expect(wrapper).toHaveStyle({
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
        });
    });
});
