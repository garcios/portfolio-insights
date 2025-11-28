import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import PortfolioChart from '../components/PortfolioChart';
import { PortfolioPerformance } from '../types/portfolio';

// Mock Recharts to avoid issues with ResponsiveContainer in JSDOM
vi.mock('recharts', () => {
    const OriginalModule = vi.importActual('recharts');
    return {
        ...OriginalModule,
        ResponsiveContainer: ({ children }: any) => <div data-testid="responsive-container">{children}</div>,
        AreaChart: ({ children }: any) => <div data-testid="area-chart">{children}</div>,
        Area: () => <div data-testid="area" />,
        XAxis: () => <div data-testid="x-axis" />,
        YAxis: () => <div data-testid="y-axis" />,
        CartesianGrid: () => <div data-testid="cartesian-grid" />,
        Tooltip: () => <div data-testid="tooltip" />,
    };
});

const mockData: PortfolioPerformance[] = [
    { date: '2023-01-01', value: 1000 },
    { date: '2023-01-02', value: 1100 },
    { date: '2023-01-03', value: 1050 },
];

describe('PortfolioChart', () => {
    it('renders chart components correctly', () => {
        render(
            <PortfolioChart
                data={mockData}
                isPositive={true}
            />
        );

        expect(screen.getByTestId('responsive-container')).toBeInTheDocument();
        expect(screen.getByTestId('area-chart')).toBeInTheDocument();
        expect(screen.getByTestId('area')).toBeInTheDocument();
        expect(screen.getByTestId('x-axis')).toBeInTheDocument();
        expect(screen.getByTestId('y-axis')).toBeInTheDocument();
        expect(screen.getByTestId('cartesian-grid')).toBeInTheDocument();
        expect(screen.getByTestId('tooltip')).toBeInTheDocument();
    });

    it('renders without crashing with empty data', () => {
        render(
            <PortfolioChart
                data={[]}
                isPositive={false}
            />
        );

        expect(screen.getByTestId('area-chart')).toBeInTheDocument();
    });
});
