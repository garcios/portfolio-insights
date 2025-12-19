import { gql } from '@apollo/client';

export const LIST_TRANSACTIONS = gql`
  query ListTransactions($pageSize: Int, $pageToken: String, $filter: TransactionFilterInput) {
    listTransactions(pageSize: $pageSize, pageToken: $pageToken, filter: $filter) {
      transactions {
        id
        userId
        symbol
        type
        quantity
        pricePerShare
        executedAt
        notes
        brokerage
        priceCurrency
        brokerageCurrency
        createdAt
        updatedAt
      }
      nextPageToken
    }
  }
`;

export const GET_USER = gql`
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      email
      username
      firstName
      lastName
      role
      preferences
      lastLoginAt
    }
  }
`;

export const GET_PORTFOLIO = gql`
  query GetPortfolio($startDate: String, $endDate: String) {
    portfolio {
    id
    userId
    name
    summary(startDate: $startDate, endDate: $endDate) {
      totalValue
      totalGainLoss
      totalGainLossPercentage
      dayChange
      dayChangePercentage
      currency
      lastUpdated
      capitalGain
      capitalGainPercentage
      currencyGain
      currencyGainPercentage
      dividends
      dividendsPercentage
    }
      holdings {
      symbol
      quantity
      averagePrice
      currentPrice
      currentValue
      gainLoss
      gainLossPercentage
      currency
      assetName
    }
  }
}
`;

export const GET_PORTFOLIO_PERFORMANCE = gql`
  query GetPortfolioPerformance($period: String!) {
  portfolioPerformance(period: $period) {
    timestamp
    value
  }
}
`;
