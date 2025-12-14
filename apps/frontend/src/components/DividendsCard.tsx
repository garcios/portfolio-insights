import React from 'react';
import { Coins } from 'lucide-react';
import StatsCard from './StatsCard';

const DividendsCard: React.FC = () => {
    // Example values as per requirements. 
    // In a real implementation, these would likely come from props or a hook.
    const dividendsAmount = 5808.52;
    const dividendsGrowth = 1.47;
    const currency = 'AUD';

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
            title="Dividends"
            value={formatCurrency(dividendsAmount)}
            change={dividendsGrowth}
            changeLabel="p.a."
            icon={Coins}
            iconColor="var(--color-success)" // Green color for icon
            valueColor="var(--color-success)" // Green color for primary metric
        />
    );
};

export default DividendsCard;
