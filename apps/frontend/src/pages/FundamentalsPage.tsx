import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, DollarSign, Activity, PieChart, ShieldCheck } from 'lucide-react';
import { mockFundamentals } from '../mocks/fundamentals';
import StatsCard from '../components/StatsCard';

// Sub-components
import CompanyInfo from '../components/fundamentals/CompanyInfo';
import ValuationRatios from '../components/fundamentals/ValuationRatios';
import ProfitabilityGrowth from '../components/fundamentals/ProfitabilityGrowth';
import FinancialHealth from '../components/fundamentals/FinancialHealth';
import DividendShareData from '../components/fundamentals/DividendShareData';
import QualitativeContext from '../components/fundamentals/QualitativeContext';

const FundamentalsPage = () => {
    const { ticker } = useParams<{ ticker: string }>();
    const company = mockFundamentals.find(f => f.ticker === ticker);

    if (!company) {
        return (
            <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
                <div style={{ padding: '40px', textAlign: 'center' }}>
                    <h2>Company not found</h2>
                    <p>The ticker {ticker} could not be found in our database.</p>
                    <Link to="/fundamentals" style={{ color: 'var(--color-primary)', marginTop: '16px', display: 'inline-block' }}>
                        Back to Screener
                    </Link>
                </div>
            </div>
        );
    }

    const formatPercent = (val: number) => `${val.toFixed(2)}%`;
    const formatLargeNumber = (val: number) => new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 2 }).format(val);

    return (
        <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>

            <main style={{ maxWidth: '1400px', margin: '0 auto', padding: '32px 24px' }}>
                {/* Breadcrumb */}
                <div style={{ marginBottom: '24px' }}>
                    <Link
                        to="/fundamentals"
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            color: 'var(--color-text-secondary)',
                            textDecoration: 'none',
                            fontSize: '0.875rem',
                            fontWeight: '500'
                        }}
                    >
                        <ArrowLeft size={16} /> Back to Screener
                    </Link>
                </div>

                {/* Core Company Info */}
                <CompanyInfo company={company} />

                {/* Key Stats Grid */}
                <div className="grid grid-cols-4" style={{ marginBottom: '32px', gap: '24px' }}>
                    <StatsCard
                        title="Market Cap"
                        value={formatLargeNumber(company.marketCap)}
                        icon={DollarSign}
                        iconColor="var(--color-primary)"
                    />
                    <StatsCard
                        title="P/E Ratio"
                        value={company.peRatio.toFixed(2)}
                        icon={Activity}
                        iconColor="var(--color-accent)"
                    />
                    <StatsCard
                        title="EPS (TTM)"
                        value={`$${company.epsTtm}`}
                        icon={PieChart}
                        iconColor="var(--color-secondary)"
                    />
                    <StatsCard
                        title="Dividend Yield"
                        value={formatPercent(company.dividendYield)}
                        icon={ShieldCheck}
                        iconColor="var(--color-success)"
                    />
                </div>

                {/* Detailed Sections */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: '24px' }}>
                    <ValuationRatios company={company} />
                    <ProfitabilityGrowth company={company} />
                    <FinancialHealth company={company} />
                    <DividendShareData company={company} />
                    <QualitativeContext company={company} />
                </div>
            </main>
        </div>
    );
};

export default FundamentalsPage;
