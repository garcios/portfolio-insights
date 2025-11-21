import React from 'react';
import { TrendingUp, TrendingDown } from 'lucide-react';
import { Holding } from '../types/portfolio';

interface HoldingsTableProps {
    holdings: Holding[];
}

const HoldingsTable: React.FC<HoldingsTableProps> = ({ holdings }) => {
    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(value);
    };

    const formatNumber = (value: number) => {
        return new Intl.NumberFormat('en-US', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(value);
    };

    // Calculate mock price per share and change for demo
    const enrichedHoldings = holdings.map(holding => {
        const pricePerShare = holding.value / holding.quantity;
        const changePercent = (Math.random() - 0.5) * 10; // Mock change
        const isPositive = changePercent >= 0;

        return {
            ...holding,
            pricePerShare,
            changePercent,
            isPositive,
        };
    });

    const totalValue = holdings.reduce((sum, h) => sum + h.value, 0);

    return (
        <div style={{ overflowX: 'auto' }}>
            <table style={{
                width: '100%',
                borderCollapse: 'separate',
                borderSpacing: '0',
            }}>
                <thead>
                    <tr style={{
                        borderBottom: '1px solid var(--color-border)',
                    }}>
                        <th style={{
                            padding: '16px 12px',
                            textAlign: 'left',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            color: 'var(--color-text-tertiary)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                        }}>
                            Symbol
                        </th>
                        <th style={{
                            padding: '16px 12px',
                            textAlign: 'right',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            color: 'var(--color-text-tertiary)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                        }}>
                            Quantity
                        </th>
                        <th style={{
                            padding: '16px 12px',
                            textAlign: 'right',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            color: 'var(--color-text-tertiary)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                        }}>
                            Price
                        </th>
                        <th style={{
                            padding: '16px 12px',
                            textAlign: 'right',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            color: 'var(--color-text-tertiary)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                        }}>
                            Change
                        </th>
                        <th style={{
                            padding: '16px 12px',
                            textAlign: 'right',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            color: 'var(--color-text-tertiary)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                        }}>
                            Value
                        </th>
                        <th style={{
                            padding: '16px 12px',
                            textAlign: 'right',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            color: 'var(--color-text-tertiary)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.05em',
                        }}>
                            Allocation
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {enrichedHoldings.map((holding, index) => {
                        const allocation = (holding.value / totalValue) * 100;

                        return (
                            <tr
                                key={holding.symbol}
                                style={{
                                    borderBottom: index < enrichedHoldings.length - 1
                                        ? '1px solid var(--color-border)'
                                        : 'none',
                                    transition: 'background-color 0.2s',
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.backgroundColor = 'var(--color-bg-hover)';
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.backgroundColor = 'transparent';
                                }}
                            >
                                <td style={{ padding: '16px 12px' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                                        <div style={{
                                            width: '40px',
                                            height: '40px',
                                            borderRadius: '50%',
                                            background: 'linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            fontSize: '0.875rem',
                                            fontWeight: '700',
                                            color: 'white',
                                        }}>
                                            {holding.symbol.substring(0, 2)}
                                        </div>
                                        <div>
                                            <div style={{
                                                fontWeight: '600',
                                                fontSize: '0.875rem',
                                                color: 'var(--color-text-primary)',
                                            }}>
                                                {holding.symbol}
                                            </div>
                                            <div style={{
                                                fontSize: '0.75rem',
                                                color: 'var(--color-text-tertiary)',
                                            }}>
                                                Stock
                                            </div>
                                        </div>
                                    </div>
                                </td>
                                <td style={{
                                    padding: '16px 12px',
                                    textAlign: 'right',
                                    fontSize: '0.875rem',
                                    fontWeight: '500',
                                    color: 'var(--color-text-secondary)',
                                }}>
                                    {formatNumber(holding.quantity)}
                                </td>
                                <td style={{
                                    padding: '16px 12px',
                                    textAlign: 'right',
                                    fontSize: '0.875rem',
                                    fontWeight: '500',
                                    color: 'var(--color-text-secondary)',
                                }}>
                                    {formatCurrency(holding.pricePerShare)}
                                </td>
                                <td style={{
                                    padding: '16px 12px',
                                    textAlign: 'right',
                                }}>
                                    <div style={{
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        gap: '4px',
                                        padding: '4px 8px',
                                        borderRadius: '6px',
                                        fontSize: '0.75rem',
                                        fontWeight: '600',
                                        backgroundColor: holding.isPositive
                                            ? 'rgba(16, 185, 129, 0.1)'
                                            : 'rgba(239, 68, 68, 0.1)',
                                        color: holding.isPositive
                                            ? 'var(--color-success)'
                                            : 'var(--color-danger)',
                                    }}>
                                        {holding.isPositive ? (
                                            <TrendingUp size={12} />
                                        ) : (
                                            <TrendingDown size={12} />
                                        )}
                                        {Math.abs(holding.changePercent).toFixed(2)}%
                                    </div>
                                </td>
                                <td style={{
                                    padding: '16px 12px',
                                    textAlign: 'right',
                                    fontSize: '0.875rem',
                                    fontWeight: '600',
                                    color: 'var(--color-text-primary)',
                                }}>
                                    {formatCurrency(holding.value)}
                                </td>
                                <td style={{
                                    padding: '16px 12px',
                                    textAlign: 'right',
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', justifyContent: 'flex-end' }}>
                                        <div style={{
                                            flex: '0 0 60px',
                                            height: '6px',
                                            backgroundColor: 'var(--color-bg-tertiary)',
                                            borderRadius: '3px',
                                            overflow: 'hidden',
                                        }}>
                                            <div style={{
                                                width: `${allocation}%`,
                                                height: '100%',
                                                background: 'linear-gradient(90deg, var(--color-primary) 0%, var(--color-secondary) 100%)',
                                                borderRadius: '3px',
                                            }} />
                                        </div>
                                        <span style={{
                                            fontSize: '0.75rem',
                                            fontWeight: '600',
                                            color: 'var(--color-text-tertiary)',
                                            minWidth: '45px',
                                            textAlign: 'right',
                                        }}>
                                            {allocation.toFixed(1)}%
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
};

export default HoldingsTable;
