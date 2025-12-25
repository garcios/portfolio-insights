import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { Allocation } from '../types/portfolio';

const COLORS = [
    '#10b981', // Emerald 500
    '#3b82f6', // Blue 500
    '#f59e0b', // Amber 500
    '#ef4444', // Red 500
    '#8b5cf6', // Violet 500
    '#ec4899', // Pink 500
    '#6366f1', // Indigo 500
    '#14b8a6', // Teal 500
];

interface AssetAllocationCardProps {
    allocations?: Allocation[];
}

const AssetAllocationCard = ({ allocations = [] }: AssetAllocationCardProps) => {

    const data = allocations.map((allocation, index) => ({
        name: allocation.symbol,
        value: allocation.percentage,
        color: COLORS[index % COLORS.length]
    })).sort((a, b) => b.value - a.value); // Sort by percentage descending

    // Handle empty state
    if (data.length === 0) {
        return (
            <div className="card fade-in" style={{
                height: '100%',
                padding: '24px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--color-text-tertiary)'
            }}>
                No allocation data available
            </div>
        );
    }

    return (
        <div className="card fade-in" style={{
            height: '100%',
            padding: '24px',
            display: 'flex',
            flexDirection: 'column'
        }}>
            <h3 style={{
                fontSize: '1.1rem',
                fontWeight: '600',
                color: 'var(--color-text-primary)',
                marginBottom: '16px'
            }}>
                Asset Allocation
            </h3>

            <div style={{ flex: 1, position: 'relative', minHeight: '200px' }}>
                <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                        <Pie
                            data={data}
                            cx="50%"
                            cy="50%"
                            innerRadius={60}
                            outerRadius={80}
                            paddingAngle={5}
                            dataKey="value"
                            stroke="none"
                        >
                            {data.map((entry, index) => (
                                <Cell key={`cell-${index}`} fill={entry.color} />
                            ))}
                        </Pie>
                        <Tooltip
                            contentStyle={{
                                backgroundColor: 'rgba(255, 255, 255, 0.1)',
                                backdropFilter: 'blur(10px)',
                                border: '1px solid rgba(255, 255, 255, 0.1)',
                                borderRadius: '8px',
                                color: 'var(--color-text-primary)'
                            }}
                            itemStyle={{ color: 'var(--color-text-primary)' }}
                            formatter={(value: number, name: string) => [`${value.toFixed(2)}%`, name]}
                        />
                        <Legend
                            verticalAlign="bottom"
                            height={36}
                            iconType="circle"
                            formatter={(value) => <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>{value}</span>}
                        />
                    </PieChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default AssetAllocationCard;
