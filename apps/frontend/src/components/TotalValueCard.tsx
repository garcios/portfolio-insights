import React from 'react';
import { DollarSign } from 'lucide-react';
import StatsCard from './StatsCard';

interface TotalValueCardProps {
    value: number;
    change: number;
    currency: string;
}

const TotalValueCard: React.FC<TotalValueCardProps> = ({ value, change, currency }) => {
    const formatCurrency = (val: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(val);
    };

    return (
        <StatsCard
            title="Total Value"
            value={formatCurrency(value)}
            change={change}
            changeLabel="All time"
            icon={DollarSign}
            iconColor="var(--color-primary)"
        />
    );
};

export default TotalValueCard;
