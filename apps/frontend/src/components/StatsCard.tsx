import React from 'react';
import { TrendingUp, TrendingDown, LucideIcon } from 'lucide-react';

interface StatsCardProps {
    title: string;
    value: string;
    change?: number;
    changeLabel?: string;
    icon: LucideIcon;
    iconColor?: string;
    valueColor?: string;
}

const StatsCard: React.FC<StatsCardProps> = ({
    title,
    value,
    change,
    changeLabel,
    icon: Icon,
    iconColor = 'var(--color-primary)',
    valueColor = 'var(--color-text-primary)',
}) => {
    const isPositive = change !== undefined && change >= 0;

    return (
        <div className="card fade-in" style={{
            position: 'relative',
            overflow: 'hidden',
        }}>
            {/* Background gradient decoration */}
            <div style={{
                position: 'absolute',
                top: '-50%',
                right: '-20%',
                width: '150px',
                height: '150px',
                background: `radial-gradient(circle, ${iconColor}20 0%, transparent 70%)`,
                pointerEvents: 'none',
            }} />

            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                position: 'relative',
            }}>
                <div style={{ flex: 1 }}>
                    <p style={{
                        fontSize: '0.875rem',
                        color: 'var(--color-text-tertiary)',
                        marginBottom: '8px',
                        fontWeight: '500',
                    }}>
                        {title}
                    </p>
                    <h3 style={{
                        fontSize: '2rem',
                        fontWeight: '700',
                        color: valueColor,
                        marginBottom: '8px',
                        lineHeight: '1',
                    }}>
                        {value}
                    </h3>
                    {change !== undefined && (
                        <div style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '6px',
                            marginTop: '12px',
                        }}>
                            <div style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px',
                                padding: '4px 10px',
                                borderRadius: '6px',
                                fontSize: '0.875rem',
                                fontWeight: '600',
                                backgroundColor: isPositive
                                    ? 'rgba(16, 185, 129, 0.15)'
                                    : 'rgba(239, 68, 68, 0.15)',
                                color: isPositive
                                    ? 'var(--color-success)'
                                    : 'var(--color-danger)',
                            }}>
                                {isPositive ? (
                                    <TrendingUp size={14} />
                                ) : (
                                    <TrendingDown size={14} />
                                )}
                                <span>{Math.abs(change).toFixed(2)}%</span>
                            </div>
                            {changeLabel && (
                                <span style={{
                                    fontSize: '0.75rem',
                                    color: 'var(--color-text-tertiary)',
                                }}>
                                    {changeLabel}
                                </span>
                            )}
                        </div>
                    )}
                </div>

                <div style={{
                    width: '56px',
                    height: '56px',
                    borderRadius: '12px',
                    background: `linear-gradient(135deg, ${iconColor}30 0%, ${iconColor}10 100%)`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                }}>
                    <Icon size={28} color={iconColor} strokeWidth={2} />
                </div>
            </div>
        </div>
    );
};

export default StatsCard;
