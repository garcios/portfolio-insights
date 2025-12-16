import React from 'react';
import { Coins } from 'lucide-react';
import StatsCard from './StatsCard';

interface DividendsCardProps {
    value: number;
    change: number;
    currency: string;
}

const DividendsCard: React.FC<DividendsCardProps> = ({ value, change, currency }) => {
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
            title="Dividends"
            value={formatCurrency(value)}
            change={change}
            changeLabel="p.a."
            icon={Coins}
            iconColor="var(--color-success)" // Green color for icon
            valueColor="var(--color-success)" // Green color for primary metric
        />
    );
};

export default DividendsCard;
