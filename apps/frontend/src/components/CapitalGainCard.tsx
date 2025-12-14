import React from 'react';
import { TrendingUp } from 'lucide-react';
import StatsCard from './StatsCard';

const CapitalGainCard: React.FC = () => {
    // Example values as per requirements.
    const capitalGainsAmount = 104093.68;
    const capitalGainsGrowth = 25.13;
    const currency = 'USD';

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(value);
    };

    return (
        <StatsCard
            title="Capital Gain"
            value={formatCurrency(capitalGainsAmount)}
            change={capitalGainsGrowth}
            changeLabel="p.a."
            icon={TrendingUp}
            iconColor="var(--color-success)" // Green color for icon
            valueColor="var(--color-success)" // Green color for primary metric
        />
    );
};

export default CapitalGainCard;
