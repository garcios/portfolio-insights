import { ApolloClient, InMemoryCache, ApolloLink } from '@apollo/client';
import { createUploadLink } from './createUploadLink';
import { getStoredTokens } from '../auth/oauth';

// Create upload link
const uploadLink = createUploadLink({
    uri: import.meta.env.VITE_GRAPHQL_URL || 'http://localhost:8080/query',
});

// Create auth link to add Authorization header
const authLink = new ApolloLink((operation, forward) => {
    // Get tokens from storage
    const tokens = getStoredTokens();

    // Add authorization header if token exists
    if (tokens?.accessToken) {
        operation.setContext({
            headers: {
                authorization: `Bearer ${tokens.accessToken}`,
            },
        });
    }

    return forward(operation);
});

export const apolloClient = new ApolloClient({
    link: authLink.concat(uploadLink),
    cache: new InMemoryCache(),
    defaultOptions: {
        watchQuery: {
            fetchPolicy: 'cache-and-network',
        },
    },
});
