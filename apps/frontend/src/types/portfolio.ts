export interface User {
    id: string;
    username: string;
    email: string;
}

export interface PortfolioSummary {
    totalValue: number;
    totalGainLoss: number;
    totalGainLossPercentage: number;
    dayChange: number;
    dayChangePercentage: number;
    currency: string;
    lastUpdated: string;
    capitalGain: number;
    capitalGainPercentage: number;
    currencyGain: number;
    currencyGainPercentage: number;
    dividends: number;
    dividendsPercentage: number;
}

export interface Holding {
    symbol: string;
    quantity: number;
    averagePrice: number;
    currentPrice: number;
    currentValue: number;
    gainLoss: number;
    gainLossPercentage: number;
    currency: string;
    targetCurrency?: string;
    currentValueInTargetCurrency?: number;
    gainLossInTargetCurrency?: number;
    assetName: string;
}

export interface Portfolio {
    id: string;
    userId: string;
    name: string;
    summary?: PortfolioSummary;
    holdings: Holding[];
}

export interface PortfolioPerformance {
    date: string;
    value: number;
}

export interface PortfolioStats {
    totalValue: number;
    totalChange: number;
    totalChangePercent: number;
    dayChange: number;
    dayChangePercent: number;
}
