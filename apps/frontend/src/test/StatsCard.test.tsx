import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import StatsCard from '../components/StatsCard';
import { Wallet } from 'lucide-react';

describe('StatsCard', () => {
    it('renders title and value correctly', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                icon={Wallet}
            />
        );

        expect(screen.getByText('Total Value')).toBeInTheDocument();
        expect(screen.getByText('$10,000')).toBeInTheDocument();
    });

    it('renders positive change correctly', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={5.25}
                changeLabel="vs last month"
                icon={Wallet}
            />
        );

        expect(screen.getByText('5.25%')).toBeInTheDocument();
        expect(screen.getByText('vs last month')).toBeInTheDocument();
        // Check for positive styling/icon if possible, or just presence
    });

    it('renders negative change correctly', () => {
        render(
            <StatsCard
                title="Total Value"
                value="$10,000"
                change={-2.5}
                icon={Wallet}
            />
        );

        expect(screen.getByText('2.50%')).toBeInTheDocument();
    });

    it('renders without change prop', () => {
        render(
            <StatsCard
                title="Holdings"
                value="5"
                icon={Wallet}
            />
        );

        expect(screen.getByText('Holdings')).toBeInTheDocument();
        expect(screen.getByText('5')).toBeInTheDocument();
        expect(screen.queryByText('%')).not.toBeInTheDocument();
    });
});
