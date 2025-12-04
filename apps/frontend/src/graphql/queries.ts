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
