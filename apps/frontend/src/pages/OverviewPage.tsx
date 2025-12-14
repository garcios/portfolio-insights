import { useState } from 'react';
import { useQuery, gql } from '@apollo/client';
import {
    TrendingUp,
    DollarSign,
} from 'lucide-react';
import StatsCard from '../components/StatsCard';
import DividendsCard from '../components/DividendsCard';
import CapitalGainsCard from '../components/CapitalGainsCard';
import PortfolioChart from '../components/PortfolioChart';
import HoldingsTable from '../components/HoldingsTable';
import LoadingSpinner from '../components/LoadingSpinner';
import Header from '../components/Header';
import AssetAllocationCard from '../components/AssetAllocationCard';
import RecentActivityCard from '../components/RecentActivityCard';
import { PortfolioPerformance } from '../types/portfolio';

const GET_PORTFOLIO = gql`
  query GetPortfolio {
    portfolio {
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
  query GetPortfolioPerformance($period: String!) {
    portfolioPerformance(period: $period) {
      timestamp
      value
    }
  }
`;

const OverviewPage = () => {
    const [selectedPeriod, setSelectedPeriod] = useState('1m');

    const { loading, error, data } = useQuery(GET_PORTFOLIO, {
        pollInterval: 30000, // Refresh every 30 seconds
    });

    const {
        loading: performanceLoading,
        data: performanceData,
    } = useQuery(GET_PORTFOLIO_PERFORMANCE, {
        variables: { period: selectedPeriod },
        pollInterval: 60000, // Refresh every 60 seconds
    });

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
                <Header />
                <LoadingSpinner />
            </div>
        );
    }

    if (error) {
        return (
            <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
                <Header />
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: 'calc(100vh - 80px)',
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
    const performance: PortfolioPerformance[] = performanceData?.portfolioPerformance?.map((point: { timestamp: string; value: number }) => ({
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
            <Header />

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
                    <CapitalGainsCard />
                    <DividendsCard />
                </div>

                {/* Chart and Activity Section */}
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: '2fr 1fr',
                    gap: '24px',
                    marginBottom: '32px',
                    alignItems: 'stretch'
                }}>
                    {/* Main Chart Column */}
                    <div className="card" style={{ padding: '24px', display: 'flex', flexDirection: 'column' }}>
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
                        <div style={{ flex: 1, minHeight: '400px' }}>
                            <PortfolioChart data={performance} isPositive={isPositive} />
                        </div>
                    </div>

                    {/* Side Cards Column */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                        <div style={{ flex: 1, minHeight: '300px' }}>
                            <AssetAllocationCard />
                        </div>
                        <div style={{ flex: 1, minHeight: '400px', maxHeight: '500px' }}>
                            <RecentActivityCard />
                        </div>
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
};

export default OverviewPage;
