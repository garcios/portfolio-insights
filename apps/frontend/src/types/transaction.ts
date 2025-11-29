export type TransactionType = 'BUY' | 'SELL' | 'SPLIT' | 'DIVIDEND';

export interface Transaction {
    id: string;
    date: string;
    ticker: string;
    type: TransactionType;
    quantity: number;
    price: number;
    currency: string;
    brokerage: number;
    total: number;
    notes?: string;
}
