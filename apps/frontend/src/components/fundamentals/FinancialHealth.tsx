import { CompanyFundamentals } from '../../types/fundamentals';

interface FinancialHealthProps {
    company: CompanyFundamentals;
}

const Row = ({ label, value }: { label: string; value: string }) => (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
        <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: '500', color: 'var(--color-text-primary)' }}>{value}</span>
    </div>
);

const FinancialHealth = ({ company }: FinancialHealthProps) => {
    const formatLargeNumber = (val: number) => {
        return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 2 }).format(val);
    };

    return (
        <div className="card" style={{ padding: '24px' }}>
            <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '20px', color: 'var(--color-text-primary)' }}>Financial Health</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Row label="Current Ratio" value={company.currentRatio.toFixed(2)} />
                <Row label="Quick Ratio" value={company.quickRatio.toFixed(2)} />
                <Row label="Debt to Equity" value={company.debtToEquity.toFixed(2)} />
                <Row label="Total Cash" value={formatLargeNumber(company.totalCash)} />
                <Row label="Total Debt" value={formatLargeNumber(company.totalLongTermDebt)} />
            </div>
        </div>
    );
};

export default FinancialHealth;
