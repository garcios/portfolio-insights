import React from 'react';
import { TrendingUp } from 'lucide-react';
import StatsCard from './StatsCard';

interface CapitalGainCardProps {
    value: number;
    change: number;
    currency: string;
}

const CapitalGainCard: React.FC<CapitalGainCardProps> = ({ value, change, currency }) => {
    const formatCurrency = (val: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(val);
    };

    return (
        <StatsCard
            title="Capital Gain"
            value={formatCurrency(value)}
            change={change}
            changeLabel="p.a."
            icon={TrendingUp}
            iconColor="var(--color-success)" // Green color for icon
            valueColor="var(--color-success)" // Green color for primary metric
        />
    );
};

export default CapitalGainCard;
