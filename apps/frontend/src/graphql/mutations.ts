import { gql } from '@apollo/client';

export const ADD_TRANSACTION = gql`
  mutation AddTransaction($input: NewTransaction!) {
    addTransaction(input: $input) {
      symbol
      type
      quantity
      pricePerShare
      priceCurrency
      executedAt
      notes
      brokerage
      brokerageCurrency
    }
  }
`;
