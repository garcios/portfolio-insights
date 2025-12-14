import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import RecentActivityCard from '../RecentActivityCard';
import { LIST_TRANSACTIONS } from '../../graphql/queries';

const mocks = [
    {
        request: {
            query: LIST_TRANSACTIONS,
            variables: {
                pageSize: 5,
            },
        },
        result: {
            data: {
                listTransactions: {
                    transactions: [
                        {
                            id: '1',
                            userId: 'user1',
                            symbol: 'AAPL',
                            type: 'BUY',
                            quantity: 10,
                            pricePerShare: 150.00,
                            executedAt: '2023-10-25T10:00:00Z',
                            priceCurrency: 'USD',
                            notes: null,
                            brokerage: 0,
                            brokerageCurrency: 'USD',
                            createdAt: '2023-10-25T10:00:00Z',
                            updatedAt: '2023-10-25T10:00:00Z',
                        },
                        {
                            id: '2',
                            userId: 'user1',
                            symbol: 'GOOGL',
                            type: 'SELL',
                            quantity: 5,
                            pricePerShare: 120.00,
                            executedAt: '2023-10-24T14:30:00Z',
                            priceCurrency: 'USD',
                            notes: null,
                            brokerage: 0,
                            brokerageCurrency: 'USD',
                            createdAt: '2023-10-24T14:30:00Z',
                            updatedAt: '2023-10-24T14:30:00Z',
                        },
                        {
                            id: '3',
                            userId: 'user1',
                            symbol: 'MSFT',
                            type: 'DIVIDEND',
                            quantity: 0,
                            pricePerShare: 0,
                            executedAt: '2023-10-23T09:00:00Z',
                            priceCurrency: 'USD',
                            total: 25.00,
                            notes: null,
                            brokerage: 0,
                            brokerageCurrency: 'USD',
                            createdAt: '2023-10-23T09:00:00Z',
                            updatedAt: '2023-10-23T09:00:00Z',
                        },
                    ],
                    nextPageToken: null,
                },
            },
        },
    },
];

describe('RecentActivityCard', () => {
    it('renders the title', async () => {
        render(
            <MockedProvider mocks={mocks} addTypename={false}>
                <RecentActivityCard />
            </MockedProvider>
        );
        expect(await screen.findByText('Recent Activity')).toBeInTheDocument();
    });

    it('renders transactions', async () => {
        render(
            <MockedProvider mocks={mocks} addTypename={false}>
                <RecentActivityCard />
            </MockedProvider>
        );

        expect(await screen.findByText('AAPL')).toBeInTheDocument();
        expect(screen.getByText('GOOGL')).toBeInTheDocument();
    });

    it('displays correct action types', async () => {
        render(
            <MockedProvider mocks={mocks} addTypename={false}>
                <RecentActivityCard />
            </MockedProvider>
        );

        expect(await screen.findByText('Bought')).toBeInTheDocument();
        expect(screen.getByText('Sold')).toBeInTheDocument();
        expect(screen.getByText('Dividend')).toBeInTheDocument();
    });
});

