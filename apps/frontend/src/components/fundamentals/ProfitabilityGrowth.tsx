import { CompanyFundamentals } from '../../types/fundamentals';

interface ProfitabilityGrowthProps {
    company: CompanyFundamentals;
}

const Row = ({ label, value, color }: { label: string; value: string; color?: string }) => (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
        <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: '500', color: color || 'var(--color-text-primary)' }}>{value}</span>
    </div>
);

const ProfitabilityGrowth = ({ company }: ProfitabilityGrowthProps) => {
    const formatPercent = (val: number) => `${val.toFixed(2)}%`;

    return (
        <div className="card" style={{ padding: '24px' }}>
            <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '20px', color: 'var(--color-text-primary)' }}>Profitability & Growth</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Row
                    label="Revenue Growth (YoY)"
                    value={formatPercent(company.revenueGrowthYoy)}
                    color={company.revenueGrowthYoy > 0 ? 'var(--color-success)' : 'var(--color-danger)'}
                />
                <Row label="Net Profit Margin" value={formatPercent(company.netProfitMargin)} />
                <Row label="Gross Margin" value={formatPercent(company.grossMargin)} />
                <Row label="Operating Margin" value={formatPercent(company.operatingMargin)} />
                <Row label="ROE" value={formatPercent(company.roe)} />
            </div>
        </div>
    );
};

export default ProfitabilityGrowth;
