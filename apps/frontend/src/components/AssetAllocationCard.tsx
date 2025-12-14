import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';

const AssetAllocationCard = () => {
    // Hardcoded data as per requirements
    // GOOGL (60%, Green), AMZN (30%, Blue/Gray), MSFT (10%, Gray)
    const data = [
        { name: 'GOOGL', value: 60, color: '#10b981' }, // Green matching portfolio line
        { name: 'AMZN', value: 30, color: '#3b82f6' }, // Blue
        { name: 'MSFT', value: 10, color: '#6b7280' },       // Gray
    ];

    return (
        <div className="card fade-in" style={{
            height: '100%',
            padding: '24px',
            display: 'flex',
            flexDirection: 'column'
        }}>
            <h3 style={{
                fontSize: '1.1rem', // Keeping consistent with StatsCard title size generally
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
                            formatter={(value: number) => [`${value}%`, '']}
                        />
                        <Legend
                            verticalAlign="bottom"
                            height={36}
                            iconType="circle"
                            formatter={(value) => <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>{value}</span>}
                        />
                    </PieChart>
                </ResponsiveContainer>

                {/* Central Text */}
                <div style={{
                    position: 'absolute',
                    top: '50%',
                    left: '50%',
                    transform: 'translate(-50%, -60%)', // Adjusted slightly up to center in the donut hole (ignoring legend)
                    textAlign: 'center',
                    pointerEvents: 'none'
                }}>
                    <div style={{
                        fontSize: '1.5rem',
                        fontWeight: '700',
                        color: 'var(--color-text-primary)'
                    }}>
                        60%
                    </div>
                </div>
            </div>
        </div>
    );
};

export default AssetAllocationCard;
