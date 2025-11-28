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
        dayChange
        dayChangePercentage
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
        assetName
      }
    }
  }
`;

const GET_PORTFOLIO_PERFORMANCE = gql`
  query GetPortfolioPerformance($userId: ID!, $period: String!) {
    portfolioPerformance(userId: $userId, period: $period) {
      timestamp
      value
    }
  }
`;

function AppContent() {
    // Hardcoded user ID for demo purposes
    const userId = "02b28ee7-9ba2-427a-b918-a3d8e2cc00dc";
    const [selectedPeriod, setSelectedPeriod] = useState('1m');
    const [isRefreshing, setIsRefreshing] = useState(false);

    const { loading, error, data, refetch } = useQuery(GET_PORTFOLIO, {
        variables: { id: userId },
        pollInterval: 30000, // Refresh every 30 seconds
    });

    const {
        loading: performanceLoading,
        error: performanceError,
        data: performanceData,
        refetch: refetchPerformance
    } = useQuery(GET_PORTFOLIO_PERFORMANCE, {
        variables: { userId, period: selectedPeriod },
        pollInterval: 60000, // Refresh every 60 seconds
    });

    const handleRefresh = async () => {
        setIsRefreshing(true);
        await Promise.all([refetch(), refetchPerformance()]);
        setIsRefreshing(false);
    };

    const handlePeriodChange = (period: string) => {
        setSelectedPeriod(period);
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
        dayChange: 0,
        dayChangePercentage: 0,
        currency: 'USD'
    };

    // Transform performance data from API
    const performance: PortfolioPerformance[] = performanceData?.portfolioPerformance?.map((point: any) => ({
        date: point.timestamp.split('T')[0], // Convert ISO timestamp to date
        value: point.value,
    })) || [];

    const isPositive = summary.totalGainLoss >= 0;



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
                        value={formatCurrency(summary.dayChange, summary.currency)}
                        change={summary.dayChangePercentage}
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
                                {selectedPeriod === '1d' && 'Last 24 hours'}
                                {selectedPeriod === '1w' && 'Last 7 days'}
                                {selectedPeriod === '1m' && 'Last 30 days'}
                                {selectedPeriod === '3m' && 'Last 3 months'}
                                {selectedPeriod === '1y' && 'Last 12 months'}
                                {selectedPeriod === 'all' && 'All time'}
                            </p>
                        </div>
                        <div style={{
                            display: 'flex',
                            gap: '8px',
                        }}>
                            {[
                                { label: '1D', value: '1d' },
                                { label: '1W', value: '1w' },
                                { label: '1M', value: '1m' },
                                { label: '3M', value: '3m' },
                                { label: '1Y', value: '1y' },
                                { label: 'ALL', value: 'all' }
                            ].map(({ label, value }) => (
                                <button
                                    key={value}
                                    onClick={() => handlePeriodChange(value)}
                                    disabled={performanceLoading}
                                    style={{
                                        padding: '6px 12px',
                                        borderRadius: '6px',
                                        fontSize: '0.75rem',
                                        fontWeight: '600',
                                        background: selectedPeriod === value
                                            ? 'var(--color-primary)'
                                            : 'var(--color-bg-tertiary)',
                                        color: selectedPeriod === value
                                            ? 'white'
                                            : 'var(--color-text-tertiary)',
                                        border: 'none',
                                        cursor: performanceLoading ? 'not-allowed' : 'pointer',
                                        transition: 'all 0.2s',
                                        opacity: performanceLoading ? 0.6 : 1,
                                    }}
                                    onMouseEnter={(e) => {
                                        if (selectedPeriod !== value && !performanceLoading) {
                                            e.currentTarget.style.background = 'var(--color-bg-hover)';
                                        }
                                    }}
                                    onMouseLeave={(e) => {
                                        if (selectedPeriod !== value) {
                                            e.currentTarget.style.background = 'var(--color-bg-tertiary)';
                                        }
                                    }}
                                >
                                    {label}
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
