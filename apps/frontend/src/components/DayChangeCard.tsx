import React from 'react';
import { TrendingUp } from 'lucide-react';
import StatsCard from './StatsCard';

interface DayChangeCardProps {
    value: number;
    change: number;
    currency: string;
}

const DayChangeCard: React.FC<DayChangeCardProps> = ({ value, change, currency }) => {
    const formatCurrency = (val: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(val);
    };

    const isPositive = value >= 0;

    return (
        <StatsCard
            title="Day Change"
            value={formatCurrency(value)}
            change={change}
            changeLabel="Today"
            icon={TrendingUp}
            iconColor={isPositive ? "var(--color-success)" : "var(--color-danger)"}
            valueColor={isPositive ? "var(--color-success)" : "var(--color-danger)"}
        />
    );
};

export default DayChangeCard;
