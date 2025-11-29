import { ApolloProvider } from '@apollo/client';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { apolloClient } from './utils/apolloClient';
import OverviewPage from './pages/OverviewPage';
import TransactionsPage from './pages/TransactionsPage';
import AuthPage from './pages/AuthPage';
import FundamentalsScreenerPage from './pages/FundamentalsScreenerPage';
import FundamentalsPage from './pages/FundamentalsPage';

function App() {
    return (
        <ApolloProvider client={apolloClient}>
            <Router>
                <Routes>
                    <Route path="/login" element={<AuthPage />} />
                    <Route path="/" element={<OverviewPage />} />
                    <Route path="/transactions" element={<TransactionsPage />} />
                    <Route path="/fundamentals" element={<FundamentalsScreenerPage />} />
                    <Route path="/fundamentals/:ticker" element={<FundamentalsPage />} />
                    {/* Add other routes as needed */}
                </Routes>
            </Router>
        </ApolloProvider>
    );
}

export default App;
