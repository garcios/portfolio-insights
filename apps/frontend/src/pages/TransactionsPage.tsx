import { useState, useMemo } from 'react';
import { Plus, Upload, Search, Filter, Download } from 'lucide-react';
import Header from '../components/Header';
import TransactionsTable from '../components/transactions/TransactionsTable';
import AddTransactionModal from '../components/transactions/AddTransactionModal';
import { mockTransactions } from '../mocks/transactions';
import { Transaction } from '../types/transaction';

const TransactionsPage = () => {
    const [transactions, setTransactions] = useState<Transaction[]>(mockTransactions);
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [sortConfig, setSortConfig] = useState<{ key: keyof Transaction; direction: 'asc' | 'desc' } | null>({ key: 'date', direction: 'desc' });
    const [filterType, setFilterType] = useState<string>('ALL');

    // Filter and Sort Transactions
    const filteredTransactions = useMemo(() => {
        let result = [...transactions];

        // Search Filter
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            result = result.filter(t =>
                t.ticker.toLowerCase().includes(query)
            );
        }

        // Type Filter
        if (filterType !== 'ALL') {
            result = result.filter(t => t.type === filterType);
        }

        // Sorting
        if (sortConfig) {
            result.sort((a, b) => {
                const aValue = a[sortConfig.key];
                const bValue = b[sortConfig.key];

                if (aValue === undefined || bValue === undefined) return 0;

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
    }, [transactions, searchQuery, sortConfig, filterType]);

    const handleSort = (key: keyof Transaction) => {
        let direction: 'asc' | 'desc' = 'asc';
        if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') {
            direction = 'desc';
        }
        setSortConfig({ key, direction });
    };

    const handleAddTransaction = (newTransaction: Omit<Transaction, 'id'>) => {
        const transaction: Transaction = {
            ...newTransaction,
            id: Math.random().toString(36).substr(2, 9), // Generate random ID
        };
        setTransactions(prev => [transaction, ...prev]);
    };

    const handleFileUpload = () => {
        // Mock file upload
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.csv';
        input.onchange = (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (file) {
                alert(`File selected: ${file.name}. CSV parsing not implemented yet.`);
            }
        };
        input.click();
    };

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

                        <div style={{ display: 'flex', gap: '12px' }}>
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

                    {/* Pagination (Mock) */}
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
                            Showing {filteredTransactions.length > 0 ? 1 : 0} to {filteredTransactions.length} of {filteredTransactions.length} results
                        </div>
                        <div style={{ display: 'flex', gap: '8px' }}>
                            <button
                                disabled
                                style={{
                                    padding: '6px 12px',
                                    borderRadius: '6px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-tertiary)',
                                    color: 'var(--color-text-tertiary)',
                                    cursor: 'not-allowed',
                                    fontSize: '0.875rem'
                                }}
                            >
                                Previous
                            </button>
                            <button
                                disabled
                                style={{
                                    padding: '6px 12px',
                                    borderRadius: '6px',
                                    border: '1px solid var(--color-border)',
                                    background: 'var(--color-bg-tertiary)',
                                    color: 'var(--color-text-tertiary)',
                                    cursor: 'not-allowed',
                                    fontSize: '0.875rem'
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
                onSave={handleAddTransaction}
            />
        </div>
    );
};

export default TransactionsPage;
