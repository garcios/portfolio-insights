import { CompanyFundamentals } from '../../types/fundamentals';

interface ValuationRatiosProps {
    company: CompanyFundamentals;
}

const Row = ({ label, value }: { label: string; value: string }) => (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem' }}>
        <span style={{ color: 'var(--color-text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: '500', color: 'var(--color-text-primary)' }}>{value}</span>
    </div>
);

const ValuationRatios = ({ company }: ValuationRatiosProps) => {
    return (
        <div className="card" style={{ padding: '24px' }}>
            <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '20px', color: 'var(--color-text-primary)' }}>Valuation</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Row label="P/E Ratio" value={company.peRatio.toFixed(2)} />
                <Row label="Forward P/E" value={company.forwardPe.toFixed(2)} />
                <Row label="PEG Ratio" value={company.pegRatio.toFixed(2)} />
                <Row label="Price to Book" value={company.priceToBook.toFixed(2)} />
                <Row label="EV/EBITDA" value={company.evToEbitda.toFixed(2)} />
            </div>
        </div>
    );
};

export default ValuationRatios;
