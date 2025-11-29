import { ApolloClient, InMemoryCache } from '@apollo/client';
import { createUploadLink } from './createUploadLink';

const uploadLink = createUploadLink({
    uri: import.meta.env.VITE_GRAPHQL_URL || 'http://localhost:8080/query',
});

export const apolloClient = new ApolloClient({
    link: uploadLink,
    cache: new InMemoryCache(),
    defaultOptions: {
        watchQuery: {
            fetchPolicy: 'cache-and-network',
        },
    },
});
