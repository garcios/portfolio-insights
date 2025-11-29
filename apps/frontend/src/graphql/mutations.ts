import { gql } from '@apollo/client';

export const CREATE_TRANSACTION = gql`
  mutation CreateTransaction($input: NewTransaction!) {
    createTransaction(input: $input) {
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

export const UPLOAD_TRANSACTION_CSV = gql`
  mutation UploadTransactionCSV($file: Upload!) {
    uploadTransactionCSV(file: $file)
  }
`;
