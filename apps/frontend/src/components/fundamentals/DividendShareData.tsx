import { CompanyFundamentals } from '../../types/fundamentals';

interface DividendShareDataProps {
    company: CompanyFundamentals;
}

const Row = ({ label, value }: { label: string; value: string }) => (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
        <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: '500', color: 'var(--color-text-primary)' }}>{value}</span>
    </div>
);

const DividendShareData = ({ company }: DividendShareDataProps) => {
    const formatPercent = (val: number) => `${val.toFixed(2)}%`;
    const formatCurrency = (val: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: company.currency }).format(val);
    const formatLargeNumber = (val: number) => new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 2 }).format(val);

    if (company.dividendYield === 0) {
        return (
            <div className="card" style={{ padding: '24px' }}>
                <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '20px', color: 'var(--color-text-primary)' }}>Dividend & Share Data</h3>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Row label="Shares Outstanding" value={formatLargeNumber(company.sharesOutstanding)} />
                    <div style={{ padding: '16px', background: 'var(--color-bg-tertiary)', borderRadius: '8px', textAlign: 'center', color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>
                        This company does not currently pay a dividend.
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="card" style={{ padding: '24px' }}>
            <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '20px', color: 'var(--color-text-primary)' }}>Dividend & Share Data</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Row label="Dividend Yield" value={formatPercent(company.dividendYield)} />
                <Row label="Annual Dividend" value={formatCurrency(company.dividendPerShare)} />
                <Row label="Payout Ratio" value={formatPercent(company.payoutRatio)} />
                <Row label="Ex-Dividend Date" value={company.exDividendDate || 'N/A'} />
                <Row label="Shares Outstanding" value={formatLargeNumber(company.sharesOutstanding)} />
            </div>
        </div>
    );
};

export default DividendShareData;
