import { ApolloProvider } from '@apollo/client';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { apolloClient } from './utils/apolloClient';
import { AuthProvider } from './auth/AuthContext';
import { ProtectedRoute } from './auth/ProtectedRoute';
import HomePage from './pages/HomePage';
import RegisterPage from './pages/RegisterPage';
import OverviewPage from './pages/OverviewPage';
import TransactionsPage from './pages/TransactionsPage';
import AuthPage from './pages/AuthPage';
import AuthCallbackPage from './pages/AuthCallbackPage';
import FundamentalsScreenerPage from './pages/FundamentalsScreenerPage';
import FundamentalsPage from './pages/FundamentalsPage';
import SettingsPage from './pages/SettingsPage';

function App() {
    return (
        <AuthProvider>
            <ApolloProvider client={apolloClient}>
                <Router>
                    <Routes>
                        {/* Public routes */}
                        <Route path="/" element={<HomePage />} />
                        <Route path="/register" element={<RegisterPage />} />
                        <Route path="/login" element={<AuthPage />} />
                        <Route path="/auth/callback" element={<AuthCallbackPage />} />

                        {/* Protected routes */}
                        <Route path="/dashboard" element={
                            <ProtectedRoute>
                                <OverviewPage />
                            </ProtectedRoute>
                        } />
                        <Route path="/transactions" element={
                            <ProtectedRoute>
                                <TransactionsPage />
                            </ProtectedRoute>
                        } />
                        <Route path="/fundamentals" element={
                            <ProtectedRoute>
                                <FundamentalsScreenerPage />
                            </ProtectedRoute>
                        } />
                        <Route path="/fundamentals/:ticker" element={
                            <ProtectedRoute>
                                <FundamentalsPage />
                            </ProtectedRoute>
                        } />
                        <Route path="/settings" element={
                            <ProtectedRoute>
                                <SettingsPage />
                            </ProtectedRoute>
                        } />
                    </Routes>
                </Router>
            </ApolloProvider>
        </AuthProvider>
    );
}

export default App;
