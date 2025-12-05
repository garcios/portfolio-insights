import React from 'react';
import {
    AreaChart,
    Area,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
} from 'recharts';
import { PortfolioPerformance } from '../types/portfolio';

interface PortfolioChartProps {
    data: PortfolioPerformance[];
    isPositive: boolean;
}

const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 0,
        maximumFractionDigits: 0,
    }).format(value);
};

const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
        return (
            <div className="glass-effect" style={{
                padding: '12px 16px',
                borderRadius: '8px',
                border: '1px solid rgba(255, 255, 255, 0.1)',
            }}>
                <p style={{
                    fontSize: '0.875rem',
                    color: 'var(--color-text-secondary)',
                    marginBottom: '4px'
                }}>
                    {formatDate(payload[0].payload.date)}
                </p>
                <p style={{
                    fontSize: '1.125rem',
                    fontWeight: '700',
                    color: 'var(--color-text-primary)'
                }}>
                    {formatCurrency(payload[0].value)}
                </p>
            </div>
        );
    }
    return null;
};

const PortfolioChart: React.FC<PortfolioChartProps> = ({ data, isPositive }) => {
    return (
        <div style={{ width: '100%', height: '100%' }}>
            <ResponsiveContainer width="100%" height="100%">
                <AreaChart
                    data={data}
                    margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
                >
                    <defs>
                        <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                            <stop
                                offset="5%"
                                stopColor={isPositive ? '#10b981' : '#ef4444'}
                                stopOpacity={0.3}
                            />
                            <stop
                                offset="95%"
                                stopColor={isPositive ? '#10b981' : '#ef4444'}
                                stopOpacity={0}
                            />
                        </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255, 255, 255, 0.05)" />
                    <XAxis
                        dataKey="date"
                        tickFormatter={formatDate}
                        stroke="var(--color-text-tertiary)"
                        style={{ fontSize: '0.75rem' }}
                        tickLine={false}
                    />
                    <YAxis
                        tickFormatter={formatCurrency}
                        stroke="var(--color-text-tertiary)"
                        style={{ fontSize: '0.75rem' }}
                        tickLine={false}
                        axisLine={false}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <Area
                        type="monotone"
                        dataKey="value"
                        stroke={isPositive ? '#10b981' : '#ef4444'}
                        strokeWidth={2}
                        fillOpacity={1}
                        fill="url(#colorValue)"
                    />
                </AreaChart>
            </ResponsiveContainer>
        </div>
    );
};

export default PortfolioChart;
