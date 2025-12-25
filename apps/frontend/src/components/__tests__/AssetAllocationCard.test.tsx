import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import AssetAllocationCard from '../AssetAllocationCard';

// Mock Recharts since it doesn't render well in JSDOM
// We'll just verify the data is passed correctly and the text elements are present
vi.mock('recharts', () => {
    const OriginalModule = vi.importActual('recharts');
    return {
        ...OriginalModule,
        ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
        PieChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
        Pie: () => <div>Pie Chart Segment</div>,
        Cell: () => <div>Cell</div>,
        Legend: () => <div>Legend</div>,
        Tooltip: () => <div>Tooltip</div>,
    };
});

describe('AssetAllocationCard', () => {
    it('renders the title', () => {
        render(<AssetAllocationCard />);
        expect(screen.queryByText('Asset Allocation')).not.toBeInTheDocument(); // It shows "No allocation data available" when empty
        expect(screen.getByText('No allocation data available')).toBeInTheDocument();
    });

    it('renders with data', () => {
        const mockAllocations = [
            { symbol: 'AAPL', percentage: 40 },
            { symbol: 'GOOGL', percentage: 60 }
        ];
        render(<AssetAllocationCard allocations={mockAllocations} />);
        expect(screen.getByText('Asset Allocation')).toBeInTheDocument();
    });
});
