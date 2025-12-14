import React from 'react';
import { ArrowRightLeft } from 'lucide-react';
import StatsCard from './StatsCard';

const CurrencyGainCard: React.FC = () => {
    // Example values as per requirements.
    const currencyGainAmount = -2091.91;
    const currencyGainGrowth = -0.53;
    const currency = 'USD';

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(value);
    };

    // Determine colors based on value (Red/Green logic)
    // Requirement mentioned "Green-associated card family" but also "Red/Green based on the value".
    // For a negative financial metric, Red is the standard convention to avoid confusion.
    const isPositive = currencyGainAmount >= 0;
    const colorVar = isPositive ? 'var(--color-success)' : 'var(--color-danger)';

    return (
        <StatsCard
            title="Currency Gain"
            value={formatCurrency(currencyGainAmount)}
            change={currencyGainGrowth}
            changeLabel="p.a."
            icon={ArrowRightLeft}
            iconColor={colorVar}
            valueColor={colorVar}
        />
    );
};

export default CurrencyGainCard;
