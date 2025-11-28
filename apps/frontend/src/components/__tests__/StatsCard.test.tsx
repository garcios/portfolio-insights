import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DollarSign, TrendingUp } from 'lucide-react';
import StatsCard from '../StatsCard';

describe('StatsCard', () => {
    it('renders title and value', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={DollarSign}
            />
        );

        expect(screen.getByText('Total Value')).toBeInTheDocument();
        expect(screen.getByText('$10,000')).toBeInTheDocument();
    });

    it('renders icon', () => {
        const { container } = render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={DollarSign}
            />
        );

        // Check that an SVG (icon) is rendered
        const svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });

    it('displays positive change with correct styling', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={5.5}
                changeLabel="today"
                icon={DollarSign}
            />
        );

        expect(screen.getByText('5.50%')).toBeInTheDocument();
        expect(screen.getByText('today')).toBeInTheDocument();

        // Check for TrendingUp icon
        const container = screen.getByText('5.50%').closest('div');
        expect(container).toBeInTheDocument();
    });

    it('displays negative change with correct styling', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={-3.2}
                changeLabel="today"
                icon={DollarSign}
            />
        );

        expect(screen.getByText('3.20%')).toBeInTheDocument();
        expect(screen.getByText('today')).toBeInTheDocument();
    });

    it('displays zero change as positive', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={0}
                icon={DollarSign}
            />
        );

        expect(screen.getByText('0.00%')).toBeInTheDocument();
    });

    it('does not display change when not provided', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={DollarSign}
            />
        );

        // Should not find any percentage text
        expect(screen.queryByText(/%/)).not.toBeInTheDocument();
    });

    it('renders without changeLabel', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={5.5}
                icon={DollarSign}
            />
        );

        expect(screen.getByText('5.50%')).toBeInTheDocument();
        // No changeLabel should be present
    });

    it('applies custom icon color', () => {
        const customColor = '#FF5733';
        const { container } = render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={DollarSign}
                iconColor={customColor}
            />
        );

        const svg = container.querySelector('svg');
        expect(svg).toHaveAttribute('stroke', customColor);
    });

    it('uses default icon color when not provided', () => {
        const { container } = render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={DollarSign}
            />
        );

        const svg = container.querySelector('svg');
        expect(svg).toHaveAttribute('stroke', 'var(--color-primary)');
    });

    it('renders with different icons', () => {
        const { rerender, container } = render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={DollarSign}
            />
        );

        let svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();

        rerender(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={TrendingUp}
            />
        );

        svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });

    it('formats large change values correctly', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={123.456789}
                icon={DollarSign}
            />
        );

        // Should be formatted to 2 decimal places
        expect(screen.getByText('123.46%')).toBeInTheDocument();
    });
});
