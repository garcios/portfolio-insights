export interface User {
    id: string;
    username: string;
    email: string;
}

export interface Holding {
    symbol: string;
    quantity: number;
    value: number;
}

export interface Portfolio {
    id: string;
    userId: string;
    name: string;
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
