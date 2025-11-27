import React from 'react';
import { TrendingUp, TrendingDown } from 'lucide-react';
import { Holding } from '../types/portfolio';

interface HoldingsTableProps {
    holdings: Holding[];
}

const HoldingsTable: React.FC<HoldingsTableProps> = ({ holdings }) => {
    const formatCurrency = (value: number, currency: string) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
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

    // Group holdings by currency
    const holdingsByCurrency = holdings.reduce((acc, holding) => {
        const currency = holding.currency || 'USD';
        if (!acc[currency]) {
            acc[currency] = [];
        }
        acc[currency].push(holding);
        return acc;
    }, {} as Record<string, Holding[]>);

    // Sort holdings by value within each currency group
    Object.keys(holdingsByCurrency).forEach(currency => {
        holdingsByCurrency[currency].sort((a, b) => b.currentValue - a.currentValue);
    });

    const currencies = Object.keys(holdingsByCurrency).sort();

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
                            Avg Price
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
                            Current Price
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
                            Gain/Loss
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
                            % Change
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
                    </tr>
                </thead>
                <tbody>
                    {currencies.map(currency => (
                        <React.Fragment key={currency}>
                            {/* Currency Header */}
                            <tr style={{ background: 'var(--color-bg-tertiary)' }}>
                                <td colSpan={7} style={{
                                    padding: '8px 12px',
                                    fontWeight: '700',
                                    fontSize: '0.875rem',
                                    color: 'var(--color-text-primary)',
                                    borderBottom: '1px solid var(--color-border)',
                                }}>
                                    {currency} Holdings
                                </td>
                            </tr>

                            {/* Holdings Rows */}
                            {holdingsByCurrency[currency].map((holding, index) => {
                                const isPositive = holding.gainLoss >= 0;
                                const isLastInGroup = index === holdingsByCurrency[currency].length - 1;

                                return (
                                    <tr
                                        key={holding.symbol}
                                        style={{
                                            borderBottom: isLastInGroup
                                                ? 'none' // Let the next group header or table end handle the border
                                                : '1px solid var(--color-border)',
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
                                                        {holding.assetName}
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
                                            {formatCurrency(holding.averagePrice, holding.currency)}
                                        </td>
                                        <td style={{
                                            padding: '16px 12px',
                                            textAlign: 'right',
                                            fontSize: '0.875rem',
                                            fontWeight: '500',
                                            color: 'var(--color-text-secondary)',
                                        }}>
                                            {formatCurrency(holding.currentPrice, holding.currency)}
                                        </td>
                                        <td style={{
                                            padding: '16px 12px',
                                            textAlign: 'right',
                                            fontSize: '0.875rem',
                                            fontWeight: '600',
                                            color: isPositive ? 'var(--color-success)' : 'var(--color-danger)',
                                        }}>
                                            {isPositive ? '+' : ''}{formatCurrency(holding.gainLoss, holding.currency)}
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
                                                backgroundColor: isPositive
                                                    ? 'rgba(16, 185, 129, 0.1)'
                                                    : 'rgba(239, 68, 68, 0.1)',
                                                color: isPositive
                                                    ? 'var(--color-success)'
                                                    : 'var(--color-danger)',
                                            }}>
                                                {isPositive ? (
                                                    <TrendingUp size={12} />
                                                ) : (
                                                    <TrendingDown size={12} />
                                                )}
                                                {Math.abs(holding.gainLossPercentage).toFixed(2)}%
                                            </div>
                                        </td>
                                        <td style={{
                                            padding: '16px 12px',
                                            textAlign: 'right',
                                            fontSize: '0.875rem',
                                            fontWeight: '600',
                                            color: 'var(--color-text-primary)',
                                        }}>
                                            {formatCurrency(holding.currentValue, holding.currency)}
                                        </td>
                                    </tr>
                                );
                            })}

                            {/* Subtotal Row */}
                            {(() => {
                                const currencyHoldings = holdingsByCurrency[currency];
                                const subtotalGainLoss = currencyHoldings.reduce((sum, h) => sum + h.gainLoss, 0);
                                const subtotalValue = currencyHoldings.reduce((sum, h) => sum + h.currentValue, 0);
                                const isSubtotalPositive = subtotalGainLoss >= 0;

                                return (
                                    <tr style={{
                                        background: 'var(--color-bg-tertiary)',
                                        borderBottom: '2px solid var(--color-border)',
                                        fontWeight: '700',
                                    }}>
                                        <td style={{
                                            padding: '12px',
                                            fontSize: '0.875rem',
                                            color: 'var(--color-text-primary)',
                                        }}>
                                            {currency} Subtotal
                                        </td>
                                        <td colSpan={3}></td>
                                        <td style={{
                                            padding: '12px',
                                            textAlign: 'right',
                                            fontSize: '0.875rem',
                                            fontWeight: '700',
                                            color: isSubtotalPositive ? 'var(--color-success)' : 'var(--color-danger)',
                                        }}>
                                            {isSubtotalPositive ? '+' : ''}{formatCurrency(subtotalGainLoss, currency)}
                                        </td>
                                        <td></td>
                                        <td style={{
                                            padding: '12px',
                                            textAlign: 'right',
                                            fontSize: '0.875rem',
                                            fontWeight: '700',
                                            color: 'var(--color-text-primary)',
                                        }}>
                                            {formatCurrency(subtotalValue, currency)}
                                        </td>
                                    </tr>
                                );
                            })()}
                        </React.Fragment>
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default HoldingsTable;
