import { CompanyFundamentals } from '../../types/fundamentals';

interface QualitativeContextProps {
    company: CompanyFundamentals;
}

const QualitativeContext = ({ company }: QualitativeContextProps) => {
    return (
        <div className="card" style={{ padding: '24px' }}>
            <h3 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '20px', color: 'var(--color-text-primary)' }}>Context</h3>

            <div style={{ marginBottom: '24px' }}>
                <h4 style={{ fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text-secondary)', marginBottom: '8px' }}>Competitive Moat</h4>
                <p style={{ fontSize: '0.875rem', color: 'var(--color-text-primary)', lineHeight: '1.5' }}>{company.moat}</p>
            </div>

            <div style={{ marginBottom: '24px' }}>
                <h4 style={{ fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text-secondary)', marginBottom: '8px' }}>Management</h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                    {company.executives.map((exec, i) => (
                        <div key={i} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.875rem' }}>
                            <span style={{ color: 'var(--color-text-primary)', fontWeight: '500' }}>{exec.name}</span>
                            <span style={{ color: 'var(--color-text-tertiary)' }}>{exec.title}</span>
                        </div>
                    ))}
                </div>
            </div>

            <div>
                <h4 style={{ fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text-secondary)', marginBottom: '12px' }}>Recent News</h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    {company.news.map((item, i) => (
                        <a key={i} href={item.url} style={{ textDecoration: 'none', display: 'block' }}>
                            <div style={{ fontSize: '0.875rem', fontWeight: '500', color: 'var(--color-primary)', marginBottom: '2px' }}>{item.title}</div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--color-text-tertiary)' }}>{item.source} • {item.date}</div>
                        </a>
                    ))}
                    {company.news.length === 0 && <span style={{ fontSize: '0.875rem', color: 'var(--color-text-tertiary)' }}>No recent news</span>}
                </div>
            </div>
        </div>
    );
};

export default QualitativeContext;
