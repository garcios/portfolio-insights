import React, { useState, useEffect } from 'react';
import { ApolloProvider } from '@apollo/client';
import {
    Wallet,
    TrendingUp,
    DollarSign,
    PieChart,
    RefreshCw,
    Settings,
    Bell,
    User,
} from 'lucide-react';
import { apolloClient } from './utils/apolloClient';
import StatsCard from './components/StatsCard';
import PortfolioChart from './components/PortfolioChart';
import HoldingsTable from './components/HoldingsTable';
import LoadingSpinner from './components/LoadingSpinner';
import { Portfolio, PortfolioPerformance, Holding } from './types/portfolio';

// Mock data generator for demonstration
const generateMockData = (): {
    portfolio: Portfolio;
    performance: PortfolioPerformance[];
    stats: {
        totalValue: number;
        totalChange: number;
        dayChange: number;
    };
} => {
    // Generate mock holdings
    const symbols = ['AAPL', 'GOOGL', 'MSFT', 'AMZN', 'TSLA', 'NVDA', 'META'];
    const holdings: Holding[] = symbols.map(symbol => ({
        symbol,
        quantity: Math.random() * 100 + 10,
        value: Math.random() * 50000 + 10000,
    }));

    const totalValue = holdings.reduce((sum, h) => sum + h.value, 0);

    // Generate mock performance data for the last 30 days
    const performance: PortfolioPerformance[] = [];
    const today = new Date();
    let currentValue = totalValue * 0.85; // Start at 85% of current value

    for (let i = 29; i >= 0; i--) {
        const date = new Date(today);
        date.setDate(date.getDate() - i);

        // Add some randomness with upward trend
        currentValue += (Math.random() - 0.4) * (totalValue * 0.02);

        performance.push({
            date: date.toISOString().split('T')[0],
            value: currentValue,
        });
    }

    // Ensure the last value matches current total
    performance[performance.length - 1].value = totalValue;

    const totalChange = ((totalValue - performance[0].value) / performance[0].value) * 100;
    const dayChange = ((totalValue - performance[performance.length - 2].value) / performance[performance.length - 2].value) * 100;

    return {
        portfolio: {
            id: 'portfolio-1',
            userId: 'user-1',
            name: 'My Investment Portfolio',
            holdings,
        },
        performance,
        stats: {
            totalValue,
            totalChange,
            dayChange,
        },
    };
};

function AppContent() {
    const [isLoading, setIsLoading] = useState(true);
    const [data, setData] = useState<ReturnType<typeof generateMockData> | null>(null);
    const [isRefreshing, setIsRefreshing] = useState(false);

    const loadData = () => {
        setIsRefreshing(true);
        // Simulate API call
        setTimeout(() => {
            setData(generateMockData());
            setIsLoading(false);
            setIsRefreshing(false);
        }, 1000);
    };

    useEffect(() => {
        loadData();
    }, []);

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(value);
    };

    if (isLoading || !data) {
        return (
            <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
                <LoadingSpinner />
            </div>
        );
    }

    const { portfolio, performance, stats } = data;
    const isPositive = stats.totalChange >= 0;

    return (
        <div style={{
            minHeight: '100vh',
            background: 'var(--color-bg-primary)',
        }}>
            {/* Header */}
            <header style={{
                background: 'var(--color-bg-secondary)',
                borderBottom: '1px solid var(--color-border)',
                position: 'sticky',
                top: 0,
                zIndex: 100,
                backdropFilter: 'blur(10px)',
            }}>
                <div style={{
                    maxWidth: '1400px',
                    margin: '0 auto',
                    padding: '16px 24px',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <div style={{
                            width: '40px',
                            height: '40px',
                            borderRadius: '10px',
                            background: 'linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                        }}>
                            <Wallet size={24} color="white" />
                        </div>
                        <div>
                            <h1 style={{
                                fontSize: '1.5rem',
                                fontWeight: '700',
                                color: 'var(--color-text-primary)',
                                lineHeight: '1',
                            }}>
                                Portfolio Insights
                            </h1>
                            <p style={{
                                fontSize: '0.75rem',
                                color: 'var(--color-text-tertiary)',
                                marginTop: '4px',
                            }}>
                                Real-time portfolio tracking
                            </p>
                        </div>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <button
                            onClick={loadData}
                            disabled={isRefreshing}
                            style={{
                                padding: '10px',
                                borderRadius: '8px',
                                background: 'var(--color-bg-tertiary)',
                                border: '1px solid var(--color-border)',
                                color: 'var(--color-text-secondary)',
                                cursor: isRefreshing ? 'not-allowed' : 'pointer',
                                transition: 'all 0.2s',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                            onMouseEnter={(e) => {
                                if (!isRefreshing) {
                                    e.currentTarget.style.background = 'var(--color-bg-hover)';
                                    e.currentTarget.style.borderColor = 'var(--color-border-light)';
                                }
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.background = 'var(--color-bg-tertiary)';
                                e.currentTarget.style.borderColor = 'var(--color-border)';
                            }}
                        >
                            <RefreshCw
                                size={18}
                                style={{
                                    animation: isRefreshing ? 'spin 1s linear infinite' : 'none',
                                }}
                            />
                        </button>
                        <button
                            style={{
                                padding: '10px',
                                borderRadius: '8px',
                                background: 'var(--color-bg-tertiary)',
                                border: '1px solid var(--color-border)',
                                color: 'var(--color-text-secondary)',
                                cursor: 'pointer',
                                transition: 'all 0.2s',
                            }}
                            onMouseEnter={(e) => {
                                e.currentTarget.style.background = 'var(--color-bg-hover)';
                                e.currentTarget.style.borderColor = 'var(--color-border-light)';
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.background = 'var(--color-bg-tertiary)';
                                e.currentTarget.style.borderColor = 'var(--color-border)';
                            }}
                        >
                            <Bell size={18} />
                        </button>
                        <button
                            style={{
                                padding: '10px',
                                borderRadius: '8px',
                                background: 'var(--color-bg-tertiary)',
                                border: '1px solid var(--color-border)',
                                color: 'var(--color-text-secondary)',
                                cursor: 'pointer',
                                transition: 'all 0.2s',
                            }}
                            onMouseEnter={(e) => {
                                e.currentTarget.style.background = 'var(--color-bg-hover)';
                                e.currentTarget.style.borderColor = 'var(--color-border-light)';
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.background = 'var(--color-bg-tertiary)';
                                e.currentTarget.style.borderColor = 'var(--color-border)';
                            }}
                        >
                            <Settings size={18} />
                        </button>
                        <div style={{
                            width: '36px',
                            height: '36px',
                            borderRadius: '50%',
                            background: 'linear-gradient(135deg, var(--color-accent) 0%, var(--color-primary) 100%)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            cursor: 'pointer',
                            border: '2px solid var(--color-border)',
                        }}>
                            <User size={18} color="white" />
                        </div>
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main style={{
                maxWidth: '1400px',
                margin: '0 auto',
                padding: '32px 24px',
            }}>
                {/* Stats Grid */}
                <div className="grid grid-cols-4" style={{ marginBottom: '32px' }}>
                    <StatsCard
                        title="Total Value"
                        value={formatCurrency(stats.totalValue)}
                        change={stats.totalChange}
                        changeLabel="All time"
                        icon={DollarSign}
                        iconColor="var(--color-primary)"
                    />
                    <StatsCard
                        title="Day Change"
                        value={formatCurrency(stats.totalValue * (stats.dayChange / 100))}
                        change={stats.dayChange}
                        changeLabel="Today"
                        icon={TrendingUp}
                        iconColor="var(--color-success)"
                    />
                    <StatsCard
                        title="Holdings"
                        value={portfolio.holdings.length.toString()}
                        icon={PieChart}
                        iconColor="var(--color-accent)"
                    />
                    <StatsCard
                        title="Portfolio"
                        value={portfolio.name}
                        icon={Wallet}
                        iconColor="var(--color-secondary)"
                    />
                </div>

                {/* Chart Section */}
                <div className="card" style={{ marginBottom: '32px', padding: '24px' }}>
                    <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        marginBottom: '24px',
                    }}>
                        <div>
                            <h2 style={{
                                fontSize: '1.25rem',
                                fontWeight: '700',
                                color: 'var(--color-text-primary)',
                                marginBottom: '4px',
                            }}>
                                Portfolio Performance
                            </h2>
                            <p style={{
                                fontSize: '0.875rem',
                                color: 'var(--color-text-tertiary)',
                            }}>
                                Last 30 days
                            </p>
                        </div>
                        <div style={{
                            display: 'flex',
                            gap: '8px',
                        }}>
                            {['1D', '1W', '1M', '3M', '1Y', 'ALL'].map((period) => (
                                <button
                                    key={period}
                                    style={{
                                        padding: '6px 12px',
                                        borderRadius: '6px',
                                        fontSize: '0.75rem',
                                        fontWeight: '600',
                                        background: period === '1M'
                                            ? 'var(--color-primary)'
                                            : 'var(--color-bg-tertiary)',
                                        color: period === '1M'
                                            ? 'white'
                                            : 'var(--color-text-tertiary)',
                                        border: 'none',
                                        cursor: 'pointer',
                                        transition: 'all 0.2s',
                                    }}
                                    onMouseEnter={(e) => {
                                        if (period !== '1M') {
                                            e.currentTarget.style.background = 'var(--color-bg-hover)';
                                        }
                                    }}
                                    onMouseLeave={(e) => {
                                        if (period !== '1M') {
                                            e.currentTarget.style.background = 'var(--color-bg-tertiary)';
                                        }
                                    }}
                                >
                                    {period}
                                </button>
                            ))}
                        </div>
                    </div>
                    <div style={{ height: '400px' }}>
                        <PortfolioChart data={performance} isPositive={isPositive} />
                    </div>
                </div>

                {/* Holdings Table */}
                <div className="card" style={{ padding: '24px' }}>
                    <div style={{ marginBottom: '24px' }}>
                        <h2 style={{
                            fontSize: '1.25rem',
                            fontWeight: '700',
                            color: 'var(--color-text-primary)',
                            marginBottom: '4px',
                        }}>
                            Holdings
                        </h2>
                        <p style={{
                            fontSize: '0.875rem',
                            color: 'var(--color-text-tertiary)',
                        }}>
                            {portfolio.holdings.length} positions
                        </p>
                    </div>
                    <HoldingsTable holdings={portfolio.holdings} />
                </div>
            </main>
        </div>
    );
}

function App() {
    return (
        <ApolloProvider client={apolloClient}>
            <AppContent />
        </ApolloProvider>
    );
}

export default App;
