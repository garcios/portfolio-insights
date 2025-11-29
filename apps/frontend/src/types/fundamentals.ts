export interface CompanyFundamentals {
    ticker: string;
    name: string;
    sector: string;
    industry: string;
    description: string;
    logoUrl?: string;

    // Market Data
    price: number;
    currency: string;
    change: number;
    changePercent: number;
    marketCap: number;
    lastUpdated: string;

    // Valuation
    peRatio: number;
    forwardPe: number;
    pegRatio: number;
    priceToBook: number;
    evToEbitda: number;

    // Growth & Profitability
    revenueGrowthYoy: number; // Percentage
    epsTtm: number;
    revenueTtm: number;
    netIncomeTtm: number;
    grossMargin: number;
    operatingMargin: number;
    netProfitMargin: number;
    roe: number;
    roa: number;

    // Financial Health
    currentRatio: number;
    quickRatio: number;
    debtToEquity: number;
    totalLongTermDebt: number;
    freeCashFlowTtm: number;
    cashFromOperationsTtm: number;
    totalCash: number;

    // Dividends
    dividendYield: number;
    dividendPerShare: number;
    payoutRatio: number;
    exDividendDate?: string;
    sharesOutstanding: number;

    // Context
    moat: string;
    executives: { name: string; title: string }[];
    filings: { type: string; date: string; url: string }[];
    news: { title: string; source: string; date: string; url: string }[];
}

export interface IndustryAverage {
    peRatio: number;
    forwardPe: number;
    pegRatio: number;
    priceToBook: number;
    evToEbitda: number;
}
