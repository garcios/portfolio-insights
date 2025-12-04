import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation } from '@apollo/client';
import { Plus, Upload, Search, Filter, Download } from 'lucide-react';
import Header from '../components/Header';
import TransactionsTable from '../components/transactions/TransactionsTable';
import AddTransactionModal from '../components/transactions/AddTransactionModal';
import DatePicker from '../components/common/DatePicker';
import { Transaction, TransactionFilterInput } from '../types/transaction';
import { LIST_TRANSACTIONS } from '../graphql/queries';
import { UPLOAD_TRANSACTION_CSV } from '../graphql/mutations';

const TransactionsPage = () => {
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [sortConfig, setSortConfig] = useState<{ key: keyof Transaction; direction: 'asc' | 'desc' } | null>({ key: 'executedAt', direction: 'desc' });
    const [filterType, setFilterType] = useState<string>('ALL');
    const [fromDate, setFromDate] = useState<string>('');
    const [toDate, setToDate] = useState<string>('');
    const [currentPageToken, setCurrentPageToken] = useState<string>('');
    const [pageTokenHistory, setPageTokenHistory] = useState<string[]>([]);
    const [nextPageToken, setNextPageToken] = useState<string>('');

    // Build GraphQL filter
    const graphqlFilter: TransactionFilterInput = useMemo(() => {
        const filter: TransactionFilterInput = {};
        if (searchQuery) {
            filter.symbol = searchQuery;
        }
        if (filterType !== 'ALL') {
            filter.type = filterType as any;
        }
        if (fromDate) {
            filter.fromExecutedAt = new Date(fromDate).toISOString();
        }
        if (toDate) {
            // Set to end of day for the "to" date
            const toDateTime = new Date(toDate);
            toDateTime.setHours(23, 59, 59, 999);
            filter.toExecutedAt = toDateTime.toISOString();
        }
        return filter;
    }, [searchQuery, filterType, fromDate, toDate]);

    // Fetch transactions from GraphQL
    const { data, loading, error, refetch } = useQuery(LIST_TRANSACTIONS, {
        variables: {
            pageSize: 25,
            pageToken: currentPageToken || undefined,
            filter: Object.keys(graphqlFilter).length > 0 ? graphqlFilter : undefined
        }
    });

    // Update nextPageToken when data changes
    useEffect(() => {
        if (data?.listTransactions?.nextPageToken) {
            setNextPageToken(data.listTransactions.nextPageToken);
        } else {
            setNextPageToken('');
        }
    }, [data]);

    // Reset pagination when filters change
    useEffect(() => {
        setCurrentPageToken('');
        setPageTokenHistory([]);
        setNextPageToken('');
    }, [graphqlFilter]);


    // Transform GraphQL data to match component expectations
    const transactions: Transaction[] = useMemo(() => {
        if (!data?.listTransactions?.transactions) return [];

        return data.listTransactions.transactions.map((tx: Transaction) => ({
            ...tx,
            // Add computed fields for backward compatibility
            date: tx.executedAt,
            ticker: tx.symbol,
            price: tx.pricePerShare,
            currency: tx.priceCurrency || 'USD',
            total: tx.quantity * tx.pricePerShare + (tx.brokerage || 0)
        }));
    }, [data]);

    // Filter and Sort Transactions (client-side for now)
    const filteredTransactions = useMemo(() => {
        let result = [...transactions];

        // Sorting
        if (sortConfig) {
            result.sort((a, b) => {
                const aValue = a[sortConfig.key];
                const bValue = b[sortConfig.key];

                if (aValue === undefined || aValue === null || bValue === undefined || bValue === null) return 0;

                if (aValue < bValue) {
                    return sortConfig.direction === 'asc' ? -1 : 1;
                }
                if (aValue > bValue) {
                    return sortConfig.direction === 'asc' ? 1 : -1;
                }
                return 0;
            });
        }

        return result;
    }, [transactions, sortConfig]);

    const handleSort = (key: keyof Transaction) => {
        let direction: 'asc' | 'desc' = 'asc';
        if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') {
            direction = 'desc';
        }
        setSortConfig({ key, direction });
    };

    const clearFilters = () => {
        setSearchQuery('');
        setFilterType('ALL');
        setFromDate('');
        setToDate('');
    };

    const loadNextPage = () => {
        if (nextPageToken) {
            // Save current page token to history before moving to next page
            setPageTokenHistory(prev => [...prev, currentPageToken]);
            setCurrentPageToken(nextPageToken);
        }
    };

    const loadPreviousPage = () => {
        if (pageTokenHistory.length > 0) {
            // Get the previous page token from history
            const previousToken = pageTokenHistory[pageTokenHistory.length - 1];
            setPageTokenHistory(prev => prev.slice(0, -1));
            setCurrentPageToken(previousToken);
        }
    };

    const [uploadTransactionCSV] = useMutation(UPLOAD_TRANSACTION_CSV);

    const handleFileUpload = () => {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.csv';
        input.onchange = async (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (file) {
                try {
                    const { data } = await uploadTransactionCSV({
                        variables: { file }
                    });
                    if (data?.uploadTransactionCSV) {
                        alert('File uploaded successfully!');
                        refetch(); // Refetch transactions after upload
                    }
                } catch (error) {
                    console.error('Error uploading file:', error);
                    alert('Failed to upload file.');
                }
            }
        };
        input.click();
    };


    // Show loading state
    if (loading) {
        return (
            <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
                <Header />
                <main style={{ maxWidth: '1400px', margin: '0 auto', padding: '32px 24px' }}>
                    <div style={{ textAlign: 'center', padding: '64px 0', color: 'var(--color-text-secondary)' }}>
                        Loading transactions...
                    </div>
                </main>
            </div>
        );
    }

    // Show error state
    if (error) {
        return (
            <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
                <Header />
                <main style={{ maxWidth: '1400px', margin: '0 auto', padding: '32px 24px' }}>
                    <div style={{ textAlign: 'center', padding: '64px 0' }}>
                        <p style={{ color: 'var(--color-danger)', marginBottom: '16px' }}>Error loading transactions</p>
                        <p style={{ color: 'var(--color-text-tertiary)', fontSize: '0.875rem' }}>{error.message}</p>
                    </div>
                </main>
            </div>
        );
    }

    return (
        <div style={{ minHeight: '100vh', background: 'var(--color-bg-primary)' }}>
            <Header />

            <main style={{ maxWidth: '1400px', margin: '0 auto', padding: '32px 24px' }}>
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '32px'
                }}>
                    <div>
                        <h1 style={{
                            fontSize: '1.875rem',
                            fontWeight: '700',
                            color: 'var(--color-text-primary)',
                            marginBottom: '8px'
                        }}>
                            Transactions
                        </h1>
                        <p style={{ color: 'var(--color-text-tertiary)' }}>
                            Manage your investment history
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: '12px' }}>
                        <button
                            onClick={handleFileUpload}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '8px',
                                padding: '10px 16px',
                                borderRadius: '8px',
                                border: '1px solid var(--color-border)',
                                background: 'var(--color-bg-card)',
                                color: 'var(--color-text-primary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                transition: 'all 0.2s'
                            }}
                        >
                            <Upload size={18} />
                            Upload CSV
                        </button>
                        <button
                            onClick={() => setIsAddModalOpen(true)}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '8px',
                                padding: '10px 16px',
                                borderRadius: '8px',
                                border: 'none',
                                background: 'var(--color-primary)',
                                color: 'white',
                                fontWeight: '500',
                                cursor: 'pointer',
                                transition: 'all 0.2s',
                                boxShadow: 'var(--shadow-md)'
                            }}
                        >
                            <Plus size={18} />
                            Add Transaction
                        </button>
                    </div>
                </div>

                <div className="card" style={{ padding: '24px' }}>
                    {/* Filters Toolbar */}
                    <div style={{
                        display: 'flex',
                        gap: '16px',
                        marginBottom: '24px',
                        flexWrap: 'wrap'
                    }}>
                        <div style={{
                            position: 'relative',
                            flex: '1',
                            minWidth: '240px'
                        }}>
                            <Search
                                size={18}
                                style={{
                                    position: 'absolute',
                                    left: '12px',
                                    top: '50%',
                                    transform: 'translateY(-50%)',
                                    color: 'var(--color-text-tertiary)'
                                }}
                            />
                            <input
                                type="text"
                                placeholder="Search by ticker..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: '10px 12px 10px 40px',
                                    borderRadius: '8px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-primary)',
                                    color: 'var(--color-text-primary)',
                                    fontSize: '0.875rem'
                                }}
                            />
                        </div>

                        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
                            <div style={{ position: 'relative' }}>
                                <select
                                    value={filterType}
                                    onChange={(e) => setFilterType(e.target.value)}
                                    style={{
                                        appearance: 'none',
                                        padding: '10px 36px 10px 16px',
                                        borderRadius: '8px',
                                        border: '1px solid var(--color-border)',
                                        background: 'var(--color-bg-primary)',
                                        color: 'var(--color-text-primary)',
                                        fontSize: '0.875rem',
                                        cursor: 'pointer',
                                        minWidth: '140px'
                                    }}
                                >
                                    <option value="ALL">All Types</option>
                                    <option value="BUY">Buy</option>
                                    <option value="SELL">Sell</option>
                                    <option value="SPLIT">Split</option>
                                    <option value="DIVIDEND">Dividend</option>
                                </select>
                                <Filter
                                    size={16}
                                    style={{
                                        position: 'absolute',
                                        right: '12px',
                                        top: '50%',
                                        transform: 'translateY(-50%)',
                                        color: 'var(--color-text-tertiary)',
                                        pointerEvents: 'none'
                                    }}
                                />
                            </div>

                            <DatePicker
                                id="from-date"
                                label="From:"
                                value={fromDate}
                                onChange={setFromDate}
                                placeholder="Start date"
                                max={toDate || undefined}
                            />

                            <DatePicker
                                id="to-date"
                                label="To:"
                                value={toDate}
                                onChange={setToDate}
                                placeholder="End date"
                                min={fromDate || undefined}
                            />

                            <button
                                onClick={clearFilters}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    padding: '10px 16px',
                                    borderRadius: '8px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-primary)',
                                    color: 'var(--color-text-secondary)',
                                    fontSize: '0.875rem',
                                    fontWeight: '500',
                                    cursor: 'pointer',
                                    transition: 'all 0.2s'
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.background = 'var(--color-bg-hover)';
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.background = 'var(--color-bg-primary)';
                                }}
                            >
                                <Filter size={16} />
                                Clear Filters
                            </button>

                            <button
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    padding: '10px 16px',
                                    borderRadius: '8px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-primary)',
                                    color: 'var(--color-text-secondary)',
                                    fontSize: '0.875rem',
                                    fontWeight: '500',
                                    cursor: 'pointer'
                                }}
                            >
                                <Download size={16} />
                                Export
                            </button>
                        </div>
                    </div>

                    <TransactionsTable
                        transactions={filteredTransactions}
                        sortConfig={sortConfig}
                        onSort={handleSort}
                    />

                    {/* Pagination */}
                    <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        marginTop: '24px',
                        paddingTop: '24px',
                        borderTop: '1px solid var(--color-border)',
                        color: 'var(--color-text-tertiary)',
                        fontSize: '0.875rem'
                    }}>
                        <div>
                            Showing {filteredTransactions.length} results {nextPageToken ? '(more available)' : ''}
                        </div>
                        <div style={{ display: 'flex', gap: '8px' }}>
                            <button
                                onClick={loadPreviousPage}
                                disabled={pageTokenHistory.length === 0}
                                style={{
                                    padding: '6px 12px',
                                    borderRadius: '6px',
                                    border: '1px solid var(--color-border)',
                                    background: pageTokenHistory.length === 0 ? 'var(--color-bg-tertiary)' : 'var(--color-bg-primary)',
                                    color: pageTokenHistory.length === 0 ? 'var(--color-text-tertiary)' : 'var(--color-text-primary)',
                                    cursor: pageTokenHistory.length === 0 ? 'not-allowed' : 'pointer',
                                    fontSize: '0.875rem',
                                    transition: 'all 0.2s'
                                }}
                                onMouseEnter={(e) => {
                                    if (pageTokenHistory.length > 0) {
                                        e.currentTarget.style.background = 'var(--color-bg-hover)';
                                    }
                                }}
                                onMouseLeave={(e) => {
                                    if (pageTokenHistory.length > 0) {
                                        e.currentTarget.style.background = 'var(--color-bg-primary)';
                                    }
                                }}
                            >
                                Previous
                            </button>
                            <button
                                onClick={loadNextPage}
                                disabled={!nextPageToken}
                                style={{
                                    padding: '6px 12px',
                                    borderRadius: '6px',
                                    border: '1px solid var(--color-border)',
                                    background: !nextPageToken ? 'var(--color-bg-tertiary)' : 'var(--color-bg-primary)',
                                    color: !nextPageToken ? 'var(--color-text-tertiary)' : 'var(--color-text-primary)',
                                    cursor: !nextPageToken ? 'not-allowed' : 'pointer',
                                    fontSize: '0.875rem',
                                    transition: 'all 0.2s'
                                }}
                                onMouseEnter={(e) => {
                                    if (nextPageToken) {
                                        e.currentTarget.style.background = 'var(--color-bg-hover)';
                                    }
                                }}
                                onMouseLeave={(e) => {
                                    if (nextPageToken) {
                                        e.currentTarget.style.background = 'var(--color-bg-primary)';
                                    }
                                }}
                            >
                                Next
                            </button>
                        </div>
                    </div>
                </div>
            </main>

            <AddTransactionModal
                isOpen={isAddModalOpen}
                onClose={() => setIsAddModalOpen(false)}
                onSuccess={() => {
                    refetch(); // Refetch transactions after adding new transaction
                }}
            />
        </div>
    );
};

export default TransactionsPage;
