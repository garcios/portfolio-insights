import { ArrowDown, ArrowUp, MoreHorizontal } from 'lucide-react';
import { Transaction } from '../../types/transaction';

interface TransactionsTableProps {
    transactions: Transaction[];
    sortConfig: { key: keyof Transaction; direction: 'asc' | 'desc' } | null;
    onSort: (key: keyof Transaction) => void;
}

const TransactionsTable = ({ transactions, sortConfig, onSort }: TransactionsTableProps) => {
    const formatCurrency = (value: number, currency: string) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
        }).format(value);
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    };

    const getTypeColor = (type: string) => {
        switch (type) {
            case 'BUY':
                return 'var(--color-success)';
            case 'SELL':
                return 'var(--color-danger)';
            case 'DIVIDEND':
                return 'var(--color-accent)';
            default:
                return 'var(--color-text-secondary)';
        }
    };

    const renderSortIcon = (key: keyof Transaction) => {
        if (sortConfig?.key !== key) return null;
        return sortConfig.direction === 'asc' ? <ArrowUp size={14} /> : <ArrowDown size={14} />;
    };

    return (
        <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.875rem' }}>
                <thead>
                    <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                        <th
                            onClick={() => onSort('date')}
                            style={{
                                padding: '12px 16px',
                                textAlign: 'left',
                                color: 'var(--color-text-secondary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                userSelect: 'none'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                                Date {renderSortIcon('date')}
                            </div>
                        </th>
                        <th
                            onClick={() => onSort('ticker')}
                            style={{
                                padding: '12px 16px',
                                textAlign: 'left',
                                color: 'var(--color-text-secondary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                userSelect: 'none'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                                Ticker {renderSortIcon('ticker')}
                            </div>
                        </th>
                        <th
                            onClick={() => onSort('type')}
                            style={{
                                padding: '12px 16px',
                                textAlign: 'left',
                                color: 'var(--color-text-secondary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                userSelect: 'none'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                                Type {renderSortIcon('type')}
                            </div>
                        </th>
                        <th
                            onClick={() => onSort('quantity')}
                            style={{
                                padding: '12px 16px',
                                textAlign: 'right',
                                color: 'var(--color-text-secondary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                userSelect: 'none'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '4px' }}>
                                Quantity {renderSortIcon('quantity')}
                            </div>
                        </th>
                        <th
                            onClick={() => onSort('price')}
                            style={{
                                padding: '12px 16px',
                                textAlign: 'right',
                                color: 'var(--color-text-secondary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                userSelect: 'none'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '4px' }}>
                                Price {renderSortIcon('price')}
                            </div>
                        </th>
                        <th
                            onClick={() => onSort('total')}
                            style={{
                                padding: '12px 16px',
                                textAlign: 'right',
                                color: 'var(--color-text-secondary)',
                                fontWeight: '500',
                                cursor: 'pointer',
                                userSelect: 'none'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '4px' }}>
                                Total {renderSortIcon('total')}
                            </div>
                        </th>
                        <th style={{ padding: '12px 16px', width: '40px' }}></th>
                    </tr>
                </thead>
                <tbody>
                    {transactions.length === 0 ? (
                        <tr>
                            <td colSpan={7} style={{ padding: '32px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
                                No transactions found
                            </td>
                        </tr>
                    ) : (
                        transactions.map((transaction) => (
                            <tr
                                key={transaction.id}
                                style={{
                                    borderBottom: '1px solid var(--color-border)',
                                    transition: 'background-color 0.2s'
                                }}
                                onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'var(--color-bg-hover)'}
                                onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                            >
                                <td style={{ padding: '12px 16px', color: 'var(--color-text-primary)' }}>
                                    {formatDate(transaction.date)}
                                </td>
                                <td style={{ padding: '12px 16px', fontWeight: '600', color: 'var(--color-text-primary)' }}>
                                    {transaction.ticker}
                                </td>
                                <td style={{ padding: '12px 16px' }}>
                                    <span style={{
                                        display: 'inline-block',
                                        padding: '4px 8px',
                                        borderRadius: '4px',
                                        fontSize: '0.75rem',
                                        fontWeight: '600',
                                        backgroundColor: `${getTypeColor(transaction.type)}20`,
                                        color: getTypeColor(transaction.type),
                                        textTransform: 'capitalize'
                                    }}>
                                        {transaction.type.toLowerCase()}
                                    </span>
                                </td>
                                <td style={{ padding: '12px 16px', textAlign: 'right', color: 'var(--color-text-primary)' }}>
                                    {transaction.quantity.toLocaleString()}
                                </td>
                                <td style={{ padding: '12px 16px', textAlign: 'right', color: 'var(--color-text-primary)' }}>
                                    {formatCurrency(transaction.price, transaction.currency)}
                                </td>
                                <td style={{ padding: '12px 16px', textAlign: 'right', fontWeight: '600', color: 'var(--color-text-primary)' }}>
                                    {formatCurrency(transaction.total, transaction.currency)}
                                </td>
                                <td style={{ padding: '12px 16px', textAlign: 'right' }}>
                                    <button
                                        style={{
                                            background: 'transparent',
                                            border: 'none',
                                            cursor: 'pointer',
                                            color: 'var(--color-text-tertiary)',
                                            padding: '4px'
                                        }}
                                        aria-label="Actions"
                                    >
                                        <MoreHorizontal size={16} />
                                    </button>
                                </td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
};

export default TransactionsTable;
