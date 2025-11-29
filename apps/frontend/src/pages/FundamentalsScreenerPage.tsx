import { useState, useMemo } from 'react';
import { ArrowDown, ArrowUp, Search, Filter } from 'lucide-react';
import { Link } from 'react-router-dom';
import Header from '../components/Header';
import { mockFundamentals } from '../mocks/fundamentals';
import { CompanyFundamentals } from '../types/fundamentals';

type SortKey = keyof CompanyFundamentals;

const FundamentalsScreenerPage = () => {
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedSector, setSelectedSector] = useState('ALL');
    const [sortConfig, setSortConfig] = useState<{ key: SortKey; direction: 'asc' | 'desc' }>({ key: 'marketCap', direction: 'desc' });

    // Filters
    const [minMarketCap, setMinMarketCap] = useState<number>(0); // In Billions
    const [minDividendYield, setMinDividendYield] = useState<number>(0);
    const [maxPeRatio, setMaxPeRatio] = useState<number>(100);

    const sectors = useMemo(() => {
        const uniqueSectors = Array.from(new Set(mockFundamentals.map(f => f.sector)));
        return ['ALL', ...uniqueSectors];
    }, []);

    const filteredData = useMemo(() => {
        return mockFundamentals.filter(item => {
            const matchesSearch =
                item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                item.ticker.toLowerCase().includes(searchQuery.toLowerCase());

            const matchesSector = selectedSector === 'ALL' || item.sector === selectedSector;

            const matchesMarketCap = (item.marketCap / 1000000000) >= minMarketCap;
            const matchesDividend = item.dividendYield >= minDividendYield;
            const matchesPe = item.peRatio <= maxPeRatio;

            return matchesSearch && matchesSector && matchesMarketCap && matchesDividend && matchesPe;
        }).sort((a, b) => {
            const aValue = a[sortConfig.key];
            const bValue = b[sortConfig.key];

            if (typeof aValue === 'string' && typeof bValue === 'string') {
                return sortConfig.direction === 'asc'
                    ? aValue.localeCompare(bValue)
                    : bValue.localeCompare(aValue);
            }

            if (typeof aValue === 'number' && typeof bValue === 'number') {
                return sortConfig.direction === 'asc' ? aValue - bValue : bValue - aValue;
            }

            return 0;
        });
    }, [searchQuery, selectedSector, sortConfig, minMarketCap, minDividendYield, maxPeRatio]);

    const handleSort = (key: SortKey) => {
        setSortConfig(current => ({
            key,
            direction: current.key === key && current.direction === 'desc' ? 'asc' : 'desc'
        }));
    };

    const renderSortIcon = (key: SortKey) => {
        if (sortConfig.key !== key) return null;
        return sortConfig.direction === 'asc' ? <ArrowUp size={14} /> : <ArrowDown size={14} />;
    };

    const formatCurrency = (val: number) => {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', notation: 'compact', maximumFractionDigits: 1 }).format(val);
    };

    const formatPercent = (val: number) => {
        return `${val.toFixed(2)}%`;
    };

    const formatLargeNumber = (val: number) => {
        return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 2 }).format(val);
    };

    return (
        <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
            <Header />

            <main style={{ maxWidth: '1400px', margin: '0 auto', padding: '32px 24px' }}>
                <div style={{ marginBottom: '32px' }}>
                    <h1 style={{ fontSize: '1.875rem', fontWeight: '700', color: 'var(--color-text-primary)', marginBottom: '8px' }}>
                        Asset Fundamentals
                    </h1>
                    <p style={{ color: 'var(--color-text-tertiary)' }}>
                        Screen and analyze companies based on key financial metrics.
                    </p>
                </div>

                {/* Filters Bar */}
                <div className="card" style={{ padding: '20px', marginBottom: '24px' }}>
                    <div style={{ display: 'flex', gap: '20px', flexWrap: 'wrap', alignItems: 'center' }}>
                        {/* Search */}
                        <div style={{ position: 'relative', flex: '1', minWidth: '250px' }}>
                            <Search size={18} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-tertiary)' }} />
                            <input
                                type="text"
                                placeholder="Search Ticker or Company..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: '10px 12px 10px 40px',
                                    borderRadius: '8px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-primary)',
                                    color: 'var(--color-text-primary)'
                                }}
                            />
                        </div>

                        {/* Sector Filter */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <Filter size={16} color="var(--color-text-secondary)" />
                            <select
                                value={selectedSector}
                                onChange={(e) => setSelectedSector(e.target.value)}
                                style={{
                                    padding: '10px',
                                    borderRadius: '8px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-primary)',
                                    color: 'var(--color-text-primary)',
                                    minWidth: '180px'
                                }}
                            >
                                {sectors.map(s => <option key={s} value={s}>{s === 'ALL' ? 'All Sectors' : s}</option>)}
                            </select>
                        </div>

                        {/* Additional Filters */}
                        <div style={{ display: 'flex', gap: '16px', alignItems: 'center', borderLeft: '1px solid var(--color-border)', paddingLeft: '16px' }}>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                                <label style={{ fontSize: '0.75rem', color: 'var(--color-text-tertiary)' }}>Min Market Cap ($B)</label>
                                <input
                                    type="number"
                                    value={minMarketCap}
                                    onChange={(e) => setMinMarketCap(Number(e.target.value))}
                                    style={{ width: '80px', padding: '6px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'var(--color-bg-primary)', color: 'var(--color-text-primary)' }}
                                />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                                <label style={{ fontSize: '0.75rem', color: 'var(--color-text-tertiary)' }}>Min Div Yield (%)</label>
                                <input
                                    type="number"
                                    value={minDividendYield}
                                    onChange={(e) => setMinDividendYield(Number(e.target.value))}
                                    style={{ width: '80px', padding: '6px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'var(--color-bg-primary)', color: 'var(--color-text-primary)' }}
                                />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                                <label style={{ fontSize: '0.75rem', color: 'var(--color-text-tertiary)' }}>Max P/E</label>
                                <input
                                    type="number"
                                    value={maxPeRatio}
                                    onChange={(e) => setMaxPeRatio(Number(e.target.value))}
                                    style={{ width: '80px', padding: '6px', borderRadius: '6px', border: '1px solid var(--color-border)', background: 'var(--color-bg-primary)', color: 'var(--color-text-primary)' }}
                                />
                            </div>
                        </div>
                    </div>
                </div>

                {/* Data Table */}
                <div className="card" style={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.875rem' }}>
                        <thead>
                            <tr style={{ borderBottom: '1px solid var(--color-border)', background: 'var(--color-bg-secondary)' }}>
                                {[
                                    { key: 'name', label: 'Company', align: 'left' },
                                    { key: 'ticker', label: 'Ticker', align: 'left' },
                                    { key: 'sector', label: 'Sector', align: 'left' },
                                    { key: 'price', label: 'Price', align: 'right' },
                                    { key: 'marketCap', label: 'Market Cap', align: 'right' },
                                    { key: 'peRatio', label: 'P/E (TTM)', align: 'right' },
                                    { key: 'forwardPe', label: 'Fwd P/E', align: 'right' },
                                    { key: 'revenueGrowthYoy', label: 'Rev Growth', align: 'right' },
                                    { key: 'netProfitMargin', label: 'Net Margin', align: 'right' },
                                    { key: 'roe', label: 'ROE', align: 'right' },
                                    { key: 'debtToEquity', label: 'D/E', align: 'right' },
                                    { key: 'dividendYield', label: 'Div Yield', align: 'right' },
                                ].map((col) => (
                                    <th
                                        key={col.key}
                                        onClick={() => handleSort(col.key as SortKey)}
                                        style={{
                                            padding: '12px 16px',
                                            textAlign: col.align as any,
                                            color: 'var(--color-text-secondary)',
                                            fontWeight: '600',
                                            cursor: 'pointer',
                                            whiteSpace: 'nowrap'
                                        }}
                                    >
                                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: col.align === 'right' ? 'flex-end' : 'flex-start', gap: '4px' }}>
                                            {col.label} {renderSortIcon(col.key as SortKey)}
                                        </div>
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {filteredData.length === 0 ? (
                                <tr>
                                    <td colSpan={12} style={{ padding: '32px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
                                        No companies found matching your criteria.
                                    </td>
                                </tr>
                            ) : (
                                filteredData.map((item) => (
                                    <tr
                                        key={item.ticker}
                                        style={{
                                            borderBottom: '1px solid var(--color-border)',
                                            transition: 'background-color 0.2s'
                                        }}
                                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'var(--color-bg-hover)'}
                                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                                    >
                                        <td style={{ padding: '12px 16px' }}>
                                            <Link
                                                to={`/fundamentals/${item.ticker}`}
                                                style={{
                                                    color: 'var(--color-text-primary)',
                                                    fontWeight: '600',
                                                    textDecoration: 'none',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    gap: '6px'
                                                }}
                                            >
                                                {item.name}
                                            </Link>
                                        </td>
                                        <td style={{ padding: '12px 16px' }}>
                                            <span style={{
                                                background: 'var(--color-bg-tertiary)',
                                                padding: '2px 6px',
                                                borderRadius: '4px',
                                                fontSize: '0.75rem',
                                                fontWeight: '600',
                                                color: 'var(--color-text-secondary)'
                                            }}>
                                                {item.ticker}
                                            </span>
                                        </td>
                                        <td style={{ padding: '12px 16px', color: 'var(--color-text-secondary)' }}>{item.sector}</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right', fontWeight: '500' }}>{formatCurrency(item.price)}</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{formatLargeNumber(item.marketCap)}</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{item.peRatio.toFixed(1)}x</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{item.forwardPe.toFixed(1)}x</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right', color: item.revenueGrowthYoy >= 0 ? 'var(--color-success)' : 'var(--color-danger)' }}>
                                            {formatPercent(item.revenueGrowthYoy)}
                                        </td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{formatPercent(item.netProfitMargin)}</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{formatPercent(item.roe)}</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{item.debtToEquity.toFixed(2)}</td>
                                        <td style={{ padding: '12px 16px', textAlign: 'right' }}>{item.dividendYield > 0 ? formatPercent(item.dividendYield) : '-'}</td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                <div style={{ marginTop: '16px', textAlign: 'right', fontSize: '0.75rem', color: 'var(--color-text-tertiary)' }}>
                    Data as of {new Date(mockFundamentals[0]?.lastUpdated || Date.now()).toLocaleString()}
                </div>
            </main>
        </div>
    );
};

export default FundamentalsScreenerPage;
