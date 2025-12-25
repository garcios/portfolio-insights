import { graphql, HttpResponse } from 'msw';
import userData from './mock-data/user.json';
import dashboardData from './mock-data/dashboard.json';
import transactionData from './mock-data/transactions.json';
import performanceData from './mock-data/performance.json';

export const handlers = [
    // Queries
    graphql.query('GetUser', () => {
        return HttpResponse.json({
            data: {
                user: userData,
            },
        });
    }),

    graphql.query('GetPortfolio', () => {
        return HttpResponse.json({
            data: {
                portfolio: dashboardData,
            },
        });
    }),

    graphql.query('ListTransactions', () => {
        return HttpResponse.json({
            data: {
                listTransactions: {
                    transactions: transactionData,
                    nextPageToken: null
                },
            },
        });
    }),

    graphql.query('GetPortfolioPerformance', () => {
        return HttpResponse.json({
            data: {
                portfolioPerformance: performanceData,
            },
        });
    }),

    // Mutations
    graphql.mutation('CreateUser', ({ variables }) => {
        return HttpResponse.json({
            data: {
                createUser: {
                    id: 'new-user-id',
                    ...variables.input,
                },
            },
        });
    }),

    graphql.mutation('UpdateUser', ({ variables }) => {
        return HttpResponse.json({
            data: {
                updateUser: {
                    id: userData.id,
                    ...variables.input,
                    email: userData.email, // Preserve unless updated
                    username: userData.username
                }
            }
        })
    }),

    graphql.mutation('CreateTransaction', ({ variables }) => {
        return HttpResponse.json({
            data: {
                createTransaction: {
                    ...variables.input,
                    executedAt: new Date().toISOString(), // Mock server time
                },
            },
        });
    }),
];
