import { Transaction } from '../types/transaction';

export const mockTransactions: Transaction[] = [
    {
        id: '1',
        date: '2023-11-15',
        ticker: 'AAPL',
        type: 'BUY',
        quantity: 10,
        price: 185.50,
        currency: 'USD',
        brokerage: 5.00,
        total: 1860.00,
        notes: 'Monthly investment'
    },
    {
        id: '2',
        date: '2023-11-10',
        ticker: 'MSFT',
        type: 'BUY',
        quantity: 5,
        price: 360.20,
        currency: 'USD',
        brokerage: 5.00,
        total: 1806.00
    },
    {
        id: '3',
        date: '2023-10-25',
        ticker: 'GOOGL',
        type: 'BUY',
        quantity: 15,
        price: 135.80,
        currency: 'USD',
        brokerage: 5.00,
        total: 2042.00
    },
    {
        id: '4',
        date: '2023-09-15',
        ticker: 'TSLA',
        type: 'SELL',
        quantity: 5,
        price: 250.00,
        currency: 'USD',
        brokerage: 5.00,
        total: 1245.00
    },
    {
        id: '5',
        date: '2023-08-01',
        ticker: 'NVDA',
        type: 'BUY',
        quantity: 8,
        price: 450.00,
        currency: 'USD',
        brokerage: 5.00,
        total: 3605.00
    },
    {
        id: '6',
        date: '2023-07-20',
        ticker: 'AMZN',
        type: 'BUY',
        quantity: 20,
        price: 130.00,
        currency: 'USD',
        brokerage: 5.00,
        total: 2605.00
    },
    {
        id: '7',
        date: '2023-06-15',
        ticker: 'AAPL',
        type: 'DIVIDEND',
        quantity: 10,
        price: 0.24,
        currency: 'USD',
        brokerage: 0,
        total: 2.40
    }
];
