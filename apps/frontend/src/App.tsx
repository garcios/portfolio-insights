import { useState } from 'react';
import { ApolloProvider, useQuery, gql } from '@apollo/client';
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
import { PortfolioPerformance } from './types/portfolio';

const GET_PORTFOLIO = gql`
  query GetPortfolio($id: ID!) {
    portfolio(id: $id) {
      id
      userId
      name
      summary {
        totalValue
        totalGainLoss
        totalGainLossPercentage
        currency
        lastUpdated
      }
      holdings {
        symbol
        quantity
        averagePrice
        currentPrice
        currentValue
        gainLoss
        gainLossPercentage
        currency
      }
    }
  }
`;

// Helper to generate mock performance data based on current value
// since the API doesn't support history yet
const generateMockPerformance = (currentValue: number): PortfolioPerformance[] => {
    const performance: PortfolioPerformance[] = [];
    const today = new Date();
    // Start at 85% of current value 30 days ago
    let value = currentValue * 0.85;

    for (let i = 29; i >= 0; i--) {
        const date = new Date(today);
        date.setDate(date.getDate() - i);

        // Add some randomness with upward trend
        value += (Math.random() - 0.4) * (currentValue * 0.02);

        performance.push({
            date: date.toISOString().split('T')[0],
            value: value,
        });
    }

    // Ensure last point matches current value
    performance[performance.length - 1].value = currentValue;
    return performance;
};

function AppContent() {
    // Hardcoded user ID for demo purposes
    const userId = "02b28ee7-9ba2-427a-b918-a3d8e2cc00dc";

    const { loading, error, data, refetch } = useQuery(GET_PORTFOLIO, {
        variables: { id: userId },
        pollInterval: 30000, // Refresh every 30 seconds
    });

    const [isRefreshing, setIsRefreshing] = useState(false);

    const handleRefresh = async () => {
        setIsRefreshing(true);
        await refetch();
        setIsRefreshing(false);
    };

    const formatCurrency = (value: number, currency: string = 'USD') => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(value);
    };

    if (loading && !data) {
        return (
            <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
                <LoadingSpinner />
            </div>
        );
    }

    if (error) {
        return (
            <div style={{
                minHeight: '100vh',
                background: 'var(--color-bg-primary)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--color-text-primary)'
            }}>
                <div style={{ textAlign: 'center' }}>
                    <h2>Error loading portfolio</h2>
                    <p>{error.message}</p>
                    <button
                        onClick={() => window.location.reload()}
                        style={{
                            marginTop: '16px',
                            padding: '8px 16px',
                            background: 'var(--color-primary)',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer'
                        }}
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    const portfolio = data?.portfolio;
    const summary = portfolio?.summary || {
        totalValue: 0,
        totalGainLoss: 0,
        totalGainLossPercentage: 0,
        currency: 'USD'
    };

    // Generate mock performance data for the chart
    const performance = generateMockPerformance(summary.totalValue);
    const isPositive = summary.totalGainLoss >= 0;

    // Calculate day change (mocked as 1/10th of total change for demo)
    const dayChange = summary.totalGainLoss / 10;

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
                            onClick={handleRefresh}
                            disabled={isRefreshing || loading}
                            style={{
                                padding: '10px',
                                borderRadius: '8px',
                                background: 'var(--color-bg-tertiary)',
                                border: '1px solid var(--color-border)',
                                color: 'var(--color-text-secondary)',
                                cursor: (isRefreshing || loading) ? 'not-allowed' : 'pointer',
                                transition: 'all 0.2s',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                            onMouseEnter={(e) => {
                                if (!isRefreshing && !loading) {
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
                                    animation: (isRefreshing || loading) ? 'spin 1s linear infinite' : 'none',
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
                        value={formatCurrency(summary.totalValue, summary.currency)}
                        change={summary.totalGainLossPercentage}
                        changeLabel="All time"
                        icon={DollarSign}
                        iconColor="var(--color-primary)"
                    />
                    <StatsCard
                        title="Day Change"
                        value={formatCurrency(dayChange, summary.currency)}
                        change={summary.totalGainLossPercentage / 10} // Mocked day change %
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
