import { TrendingUp, TrendingDown } from 'lucide-react';
import { CompanyFundamentals } from '../../types/fundamentals';

interface CompanyInfoProps {
    company: CompanyFundamentals;
}

const CompanyInfo = ({ company }: CompanyInfoProps) => {
    const formatCurrency = (val: number) => {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: company.currency }).format(val);
    };

    return (
        <div style={{ marginBottom: '32px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '24px' }}>
                <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '8px' }}>
                        <h1 style={{ fontSize: '2rem', fontWeight: '700', color: 'var(--color-text-primary)', lineHeight: '1' }}>
                            {company.name}
                        </h1>
                        <span style={{
                            background: 'var(--color-bg-tertiary)',
                            padding: '4px 8px',
                            borderRadius: '6px',
                            fontSize: '1rem',
                            fontWeight: '600',
                            color: 'var(--color-text-secondary)'
                        }}>
                            {company.ticker}
                        </span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>
                        <span>{company.sector}</span>
                        <span>•</span>
                        <span>{company.industry}</span>
                    </div>
                </div>

                <div style={{ textAlign: 'right' }}>
                    <div style={{ fontSize: '2rem', fontWeight: '700', color: 'var(--color-text-primary)' }}>
                        {formatCurrency(company.price)}
                    </div>
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'flex-end',
                        gap: '8px',
                        color: company.change >= 0 ? 'var(--color-success)' : 'var(--color-danger)',
                        fontWeight: '600'
                    }}>
                        {company.change >= 0 ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
                        {company.change > 0 ? '+' : ''}{company.change} ({company.changePercent}%)
                    </div>
                </div>
            </div>

            <p style={{ marginTop: '16px', color: 'var(--color-text-secondary)', maxWidth: '800px', lineHeight: '1.6' }}>
                {company.description}
            </p>
        </div>
    );
};

export default CompanyInfo;
