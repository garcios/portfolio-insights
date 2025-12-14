import { useQuery } from '@apollo/client';
import { ArrowUpRight, ArrowDownLeft, DollarSign } from 'lucide-react';
import { LIST_TRANSACTIONS } from '../graphql/queries';
import { Transaction } from '../types/transaction';
import LoadingSpinner from './LoadingSpinner';

const RecentActivityCard = () => {
    const { data, loading, error } = useQuery(LIST_TRANSACTIONS, {
        variables: {
            pageSize: 5, // We only want the recent 5
        },
        pollInterval: 30000,
    });

    if (loading) return (
        <div className="card" style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <LoadingSpinner />
        </div>
    );

    if (error) return (
        <div className="card" style={{ height: '100%', padding: '24px', color: 'var(--color-danger)' }}>
            Error loading activity
        </div>
    );

    const transactions = data?.listTransactions?.transactions || [];

    const formatCurrency = (value: number, currency = 'USD') => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(value);
    };

    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    };

    const getTransactionLabel = (type: string) => {
        switch (type) {
            case 'BUY': return 'Bought';
            case 'SELL': return 'Sold';
            case 'DIVIDEND': return 'Dividend';
            default: return type;
        }
    };

    const isPositiveTransaction = (type: string) => {
        return ['SELL', 'DIVIDEND'].includes(type);
    };

    const getTransactionIcon = (type: string) => {
        if (type === 'BUY') return <ArrowDownLeft size={16} />;
        if (type === 'SELL') return <ArrowUpRight size={16} />;
        if (type === 'DIVIDEND') return <DollarSign size={16} />;
        return <ArrowUpRight size={16} />;
    };

    return (
        <div className="card fade-in" style={{
            height: '100%',
            padding: '24px',
            display: 'flex',
            flexDirection: 'column'
        }}>
            <h3 style={{
                fontSize: '1.1rem',
                fontWeight: '600',
                color: 'var(--color-text-primary)',
                marginBottom: '16px'
            }}>
                Recent Activity
            </h3>

            <div style={{ flex: 1, overflowY: 'auto' }}>
                {transactions.length === 0 ? (
                    <div style={{ color: 'var(--color-text-tertiary)', textAlign: 'center', marginTop: '20px' }}>
                        No recent activity
                    </div>
                ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                        {transactions.map((tx: Transaction) => {
                            const isPositive = isPositiveTransaction(tx.type);
                            // Calculate total if not present (for buy/sell)
                            // Note: Backend might provide raw fields, we compute total for display if needed
                            // For Dividend, we might need a different field but assuming standard calculation or pre-computed total
                            const totalValue = tx.quantity * tx.pricePerShare;

                            return (
                                <div key={tx.id} style={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center',
                                    paddingBottom: '12px',
                                    borderBottom: '1px solid var(--color-border)',
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                                        <div style={{
                                            width: '32px',
                                            height: '32px',
                                            borderRadius: '8px',
                                            background: isPositive ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)',
                                            color: isPositive ? 'var(--color-success)' : 'var(--color-danger)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center'
                                        }}>
                                            {getTransactionIcon(tx.type)}
                                        </div>
                                        <div>
                                            <div style={{
                                                fontSize: '0.875rem',
                                                fontWeight: '500',
                                                color: 'var(--color-text-primary)'
                                            }}>
                                                {getTransactionLabel(tx.type)} <span style={{ fontWeight: '700' }}>{tx.symbol}</span>
                                            </div>
                                            <div style={{
                                                fontSize: '0.75rem',
                                                color: 'var(--color-text-tertiary)'
                                            }}>
                                                {formatDate(tx.executedAt)}
                                            </div>
                                        </div>
                                    </div>
                                    <div style={{
                                        fontSize: '0.875rem',
                                        fontWeight: '600',
                                        color: isPositive ? 'var(--color-success)' : 'var(--color-text-primary)'
                                    }}>
                                        {isPositive ? '+' : ''}{formatCurrency(totalValue, tx.priceCurrency || 'USD')}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
};

export default RecentActivityCard;
