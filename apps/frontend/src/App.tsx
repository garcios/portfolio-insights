import { ApolloProvider } from '@apollo/client';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { apolloClient } from './utils/apolloClient';
import OverviewPage from './pages/OverviewPage';
import TransactionsPage from './pages/TransactionsPage';
import AuthPage from './pages/AuthPage';

function App() {
    return (
        <ApolloProvider client={apolloClient}>
            <Router>
                <Routes>
                    <Route path="/login" element={<AuthPage />} />
                    <Route path="/" element={<OverviewPage />} />
                    <Route path="/transactions" element={<TransactionsPage />} />
                    {/* Add other routes as needed */}
                </Routes>
            </Router>
        </ApolloProvider>
    );
}

export default App;
