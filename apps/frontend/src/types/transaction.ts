export type TransactionType = 'BUY' | 'SELL' | 'SPLIT' | 'DIVIDEND';

export interface Transaction {
    id: string;
    userId: string;
    symbol: string;
    type: TransactionType;
    quantity: number;
    pricePerShare: number;
    executedAt: string;
    notes?: string | null;
    brokerage?: number | null;
    priceCurrency?: string | null;
    brokerageCurrency?: string | null;
    createdAt: string;
    updatedAt: string;
    // Computed fields for display
    date?: string;
    ticker?: string;
    price?: number;
    currency?: string;
    total?: number;
}

export interface TransactionFilterInput {
    symbol?: string;
    type?: TransactionType;
    fromExecutedAt?: string;
    toExecutedAt?: string;
}

export interface TransactionConnection {
    transactions: Transaction[];
    nextPageToken?: string | null;
}
