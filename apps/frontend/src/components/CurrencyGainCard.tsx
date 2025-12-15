import React from 'react';
import { ArrowRightLeft } from 'lucide-react';
import StatsCard from './StatsCard';

interface CurrencyGainCardProps {
    value: number;
    change: number;
    currency: string;
}

const CurrencyGainCard: React.FC<CurrencyGainCardProps> = ({ value, change, currency }) => {
    const formatCurrency = (val: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(val);
    };

    // Determine colors based on value (Red/Green logic)
    // Requirement mentioned "Green-associated card family" but also "Red/Green based on the value".
    // For a negative financial metric, Red is the standard convention to avoid confusion.
    const isPositive = value >= 0;
    const colorVar = isPositive ? 'var(--color-success)' : 'var(--color-danger)';

    return (
        <StatsCard
            title="Currency Gain"
            value={formatCurrency(value)}
            change={change}
            changeLabel="p.a."
            icon={ArrowRightLeft}
            iconColor={colorVar}
            valueColor={colorVar}
        />
    );
};

export default CurrencyGainCard;
