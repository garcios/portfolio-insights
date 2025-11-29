import { CompanyFundamentals } from '../types/fundamentals';

export const mockFundamentals: CompanyFundamentals[] = [
    {
        ticker: 'AMZN',
        name: 'Amazon.com Inc.',
        sector: 'Consumer Cyclical',
        industry: 'Internet Retail',
        description: 'Amazon.com, Inc. engages in the retail sale of consumer products and subscriptions in North America and internationally. The company operates through three segments: North America, International, and Amazon Web Services (AWS).',
        price: 145.23,
        currency: 'USD',
        change: 2.45,
        changePercent: 1.72,
        marketCap: 1520000000000,
        lastUpdated: '2025-11-29T12:00:00Z',

        peRatio: 62.4,
        forwardPe: 45.2,
        pegRatio: 1.8,
        priceToBook: 8.5,
        evToEbitda: 22.1,

        revenueGrowthYoy: 12.5,
        epsTtm: 2.32,
        revenueTtm: 554000000000,
        netIncomeTtm: 24500000000,
        grossMargin: 45.2,
        operatingMargin: 5.8,
        netProfitMargin: 4.4,
        roe: 15.2,
        roa: 5.1,

        currentRatio: 1.1,
        quickRatio: 0.9,
        debtToEquity: 0.65,
        totalLongTermDebt: 65000000000,
        freeCashFlowTtm: 18500000000,
        cashFromOperationsTtm: 42000000000,
        totalCash: 54000000000,

        dividendYield: 0,
        dividendPerShare: 0,
        payoutRatio: 0,
        sharesOutstanding: 10300000000,

        moat: 'Strong network effect and cost advantages through AWS and logistics infrastructure.',
        executives: [
            { name: 'Andy Jassy', title: 'President and CEO' },
            { name: 'Brian Olsavsky', title: 'SVP and CFO' }
        ],
        filings: [
            { type: '10-K', date: '2024-02-02', url: '#' },
            { type: '10-Q', date: '2024-10-26', url: '#' }
        ],
        news: [
            { title: 'Amazon AWS revenue grows 12% YoY', source: 'Financial Times', date: '2025-11-28', url: '#' },
            { title: 'Amazon announces new logistics hubs', source: 'Reuters', date: '2025-11-25', url: '#' },
            { title: 'Analyst upgrades AMZN price target', source: 'Bloomberg', date: '2025-11-20', url: '#' }
        ]
    },
    {
        ticker: 'AAPL',
        name: 'Apple Inc.',
        sector: 'Technology',
        industry: 'Consumer Electronics',
        description: 'Apple Inc. designs, manufactures, and markets smartphones, personal computers, tablets, wearables, and accessories worldwide.',
        price: 189.50,
        currency: 'USD',
        change: -0.45,
        changePercent: -0.24,
        marketCap: 2950000000000,
        lastUpdated: '2025-11-29T12:00:00Z',

        peRatio: 28.5,
        forwardPe: 26.1,
        pegRatio: 2.4,
        priceToBook: 42.1,
        evToEbitda: 21.5,

        revenueGrowthYoy: -1.2,
        epsTtm: 6.45,
        revenueTtm: 383000000000,
        netIncomeTtm: 97000000000,
        grossMargin: 44.1,
        operatingMargin: 30.2,
        netProfitMargin: 25.3,
        roe: 160.5,
        roa: 28.4,

        currentRatio: 0.95,
        quickRatio: 0.85,
        debtToEquity: 1.8,
        totalLongTermDebt: 95000000000,
        freeCashFlowTtm: 85000000000,
        cashFromOperationsTtm: 10500000000,
        totalCash: 61000000000,

        dividendYield: 0.52,
        dividendPerShare: 0.96,
        payoutRatio: 15.2,
        exDividendDate: '2025-11-10',
        sharesOutstanding: 15600000000,

        moat: 'High switching costs and strong brand ecosystem.',
        executives: [
            { name: 'Tim Cook', title: 'CEO' },
            { name: 'Luca Maestri', title: 'CFO' }
        ],
        filings: [
            { type: '10-K', date: '2024-11-03', url: '#' },
            { type: '10-Q', date: '2024-08-03', url: '#' }
        ],
        news: [
            { title: 'Apple Vision Pro sales exceed expectations', source: 'TechCrunch', date: '2025-11-27', url: '#' },
            { title: 'Apple services revenue hits all-time high', source: 'CNBC', date: '2025-11-22', url: '#' }
        ]
    },
    {
        ticker: 'MSFT',
        name: 'Microsoft Corporation',
        sector: 'Technology',
        industry: 'Software - Infrastructure',
        description: 'Microsoft Corporation develops, licenses, and supports software, services, devices, and solutions worldwide.',
        price: 378.85,
        currency: 'USD',
        change: 3.15,
        changePercent: 0.84,
        marketCap: 2820000000000,
        lastUpdated: '2025-11-29T12:00:00Z',

        peRatio: 35.2,
        forwardPe: 30.5,
        pegRatio: 2.1,
        priceToBook: 12.5,
        evToEbitda: 24.8,

        revenueGrowthYoy: 13.0,
        epsTtm: 10.85,
        revenueTtm: 225000000000,
        netIncomeTtm: 82000000000,
        grossMargin: 69.5,
        operatingMargin: 42.1,
        netProfitMargin: 36.4,
        roe: 39.5,
        roa: 19.2,

        currentRatio: 1.8,
        quickRatio: 1.6,
        debtToEquity: 0.45,
        totalLongTermDebt: 42000000000,
        freeCashFlowTtm: 65000000000,
        cashFromOperationsTtm: 95000000000,
        totalCash: 111000000000,

        dividendYield: 0.81,
        dividendPerShare: 3.00,
        payoutRatio: 27.5,
        exDividendDate: '2025-11-15',
        sharesOutstanding: 7430000000,

        moat: 'High switching costs in enterprise software and cloud network effects.',
        executives: [
            { name: 'Satya Nadella', title: 'Chairman and CEO' },
            { name: 'Amy Hood', title: 'EVP and CFO' }
        ],
        filings: [
            { type: '10-K', date: '2024-07-27', url: '#' },
            { type: '10-Q', date: '2024-10-24', url: '#' }
        ],
        news: [
            { title: 'Microsoft Azure gains market share', source: 'WSJ', date: '2025-11-26', url: '#' }
        ]
    },
    {
        ticker: 'JPM',
        name: 'JPMorgan Chase & Co.',
        sector: 'Financial Services',
        industry: 'Banks - Diversified',
        description: 'JPMorgan Chase & Co. operates as a financial services company worldwide.',
        price: 156.20,
        currency: 'USD',
        change: 1.20,
        changePercent: 0.77,
        marketCap: 450000000000,
        lastUpdated: '2025-11-29T12:00:00Z',

        peRatio: 10.5,
        forwardPe: 10.1,
        pegRatio: 1.5,
        priceToBook: 1.6,
        evToEbitda: 8.5,

        revenueGrowthYoy: 8.5,
        epsTtm: 14.85,
        revenueTtm: 155000000000,
        netIncomeTtm: 48000000000,
        grossMargin: 0, // Not applicable for banks usually
        operatingMargin: 35.2,
        netProfitMargin: 31.0,
        roe: 16.5,
        roa: 1.2,

        currentRatio: 0, // Not applicable
        quickRatio: 0,
        debtToEquity: 1.2,
        totalLongTermDebt: 280000000000,
        freeCashFlowTtm: 0,
        cashFromOperationsTtm: 35000000000,
        totalCash: 1400000000000,

        dividendYield: 2.65,
        dividendPerShare: 4.20,
        payoutRatio: 28.2,
        exDividendDate: '2025-10-05',
        sharesOutstanding: 2900000000,

        moat: 'Cost advantage and switching costs in banking.',
        executives: [
            { name: 'Jamie Dimon', title: 'Chairman and CEO' },
            { name: 'Jeremy Barnum', title: 'CFO' }
        ],
        filings: [
            { type: '10-K', date: '2024-02-20', url: '#' }
        ],
        news: []
    },
    {
        ticker: 'TSLA',
        name: 'Tesla, Inc.',
        sector: 'Consumer Cyclical',
        industry: 'Auto Manufacturers',
        description: 'Tesla, Inc. designs, develops, manufactures, leases, and sells electric vehicles, and energy generation and storage systems.',
        price: 245.60,
        currency: 'USD',
        change: -5.40,
        changePercent: -2.15,
        marketCap: 780000000000,
        lastUpdated: '2025-11-29T12:00:00Z',

        peRatio: 75.2,
        forwardPe: 60.5,
        pegRatio: 3.5,
        priceToBook: 14.5,
        evToEbitda: 45.2,

        revenueGrowthYoy: 15.0,
        epsTtm: 3.25,
        revenueTtm: 95000000000,
        netIncomeTtm: 10500000000,
        grossMargin: 18.2,
        operatingMargin: 10.5,
        netProfitMargin: 11.0,
        roe: 22.5,
        roa: 10.2,

        currentRatio: 1.6,
        quickRatio: 1.1,
        debtToEquity: 0.15,
        totalLongTermDebt: 5000000000,
        freeCashFlowTtm: 4500000000,
        cashFromOperationsTtm: 12500000000,
        totalCash: 26000000000,

        dividendYield: 0,
        dividendPerShare: 0,
        payoutRatio: 0,
        sharesOutstanding: 3180000000,

        moat: 'Brand and scale in EV manufacturing.',
        executives: [
            { name: 'Elon Musk', title: 'CEO' },
            { name: 'Vaibhav Taneja', title: 'CFO' }
        ],
        filings: [],
        news: []
    }
];
