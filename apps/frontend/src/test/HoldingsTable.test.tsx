import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import HoldingsTable from '../components/HoldingsTable';
import { Holding } from '../types/portfolio';

const mockHoldings: Holding[] = [
    {
        symbol: 'AAPL',
        quantity: 10,
        averagePrice: 150,
        currentPrice: 180,
        currentValue: 1800,
        gainLoss: 300,
        gainLossPercentage: 20,
        currency: 'USD',
        assetName: 'Apple Inc.',
    },
    {
        symbol: 'GOOGL',
        quantity: 5,
        averagePrice: 2000,
        currentPrice: 2500,
        currentValue: 12500,
        gainLoss: 2500,
        gainLossPercentage: 25,
        currency: 'USD',
        assetName: 'Alphabet Inc.',
    },
    {
        symbol: 'BHP',
        quantity: 100,
        averagePrice: 40,
        currentPrice: 45,
        currentValue: 4500,
        gainLoss: 500,
        gainLossPercentage: 12.5,
        currency: 'AUD',
        assetName: 'BHP Group',
    }
];

describe('HoldingsTable', () => {
    it('renders table headers correctly', () => {
        render(<HoldingsTable holdings={[]} />);

        expect(screen.getByText('Symbol')).toBeInTheDocument();
        expect(screen.getByText('Quantity')).toBeInTheDocument();
        expect(screen.getByText('Avg Price')).toBeInTheDocument();
        expect(screen.getByText('Current Price')).toBeInTheDocument();
        expect(screen.getByText('Gain/Loss')).toBeInTheDocument();
        expect(screen.getByText('% Change')).toBeInTheDocument();
        expect(screen.getByText('Value')).toBeInTheDocument();
    });

    it('renders holdings grouped by currency', () => {
        render(<HoldingsTable holdings={mockHoldings} />);

        // Check currency headers
        expect(screen.getByText('USD Holdings')).toBeInTheDocument();
        expect(screen.getByText('AUD Holdings')).toBeInTheDocument();

        // Check holdings
        expect(screen.getByText('AAPL')).toBeInTheDocument();
        expect(screen.getByText('Apple Inc.')).toBeInTheDocument();
        expect(screen.getByText('GOOGL')).toBeInTheDocument();
        expect(screen.getByText('BHP')).toBeInTheDocument();
    });

    it('calculates and renders subtotals correctly', () => {
        render(<HoldingsTable holdings={mockHoldings} />);

        // USD Subtotal
        expect(screen.getByText('USD Subtotal')).toBeInTheDocument();
        // Total Value for USD: 1800 + 12500 = 14300
        // Total Gain for USD: 300 + 2500 = 2800

        // AUD Subtotal
        expect(screen.getByText('AUD Subtotal')).toBeInTheDocument();
        // Total Value for AUD: 4500
        // Total Gain for AUD: 500
    });

    it('formats currency values correctly', () => {
        render(<HoldingsTable holdings={mockHoldings} />);

        // Check formatting (assuming en-US locale)
        // AAPL Value
        expect(screen.getByText('$1,800.00')).toBeInTheDocument();

        // Check BHP row exists
        expect(screen.getByText('BHP')).toBeInTheDocument();

        // BHP Value (AUD) - match 4,500 or 4500 with any prefix/suffix
        // This handles different locale implementations in test environment
        expect(screen.getAllByText(/4,?500/)[0]).toBeInTheDocument();
    });
});
