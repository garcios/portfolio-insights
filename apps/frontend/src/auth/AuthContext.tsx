import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import {
    AuthTokens,
    getStoredTokens,
    storeTokens,
    isTokenExpired,
    refreshAccessToken,
    logout as oauthLogout,
    buildAuthorizationURL,
    exchangeCodeForTokens,
    decodeJWT,
} from './oauth';

interface User {
    id: string;
    email: string;
    username?: string;
    firstName?: string;
    lastName?: string;
    role?: string;
}

interface AuthContextType {
    user: User | null;
    tokens: AuthTokens | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    login: () => Promise<void>;
    logout: () => void;
    handleCallback: (code: string, state: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
    children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [user, setUser] = useState<User | null>(null);
    const [tokens, setTokens] = useState<AuthTokens | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    const extractUser = (authTokens: AuthTokens) => {
        if (!authTokens.idToken) return;

        const decoded = decodeJWT(authTokens.idToken);
        if (decoded) {
            setUser({
                id: decoded.sub,
                email: decoded.email,
                username: decoded.username,
                firstName: decoded.first_name,
                lastName: decoded.last_name,
                role: decoded.role,
            });
        }
    };

    const logout = () => {
        setUser(null);
        setTokens(null);
        oauthLogout();
    };

    const login = async () => {
        if (import.meta.env.MODE === 'mock') {
            const mockUser = {
                id: '1',
                email: 'user@insights.com',
                username: 'insights-user',
                firstName: 'Insights',
                lastName: 'User',
            };
            const mockTokens = {
                accessToken: 'mock-access-token',
                idToken: 'mock-id-token',
                expiresIn: 3600,
                expiresAt: Date.now() + 3600 * 1000,
                tokenType: 'Bearer',
            };

            setUser(mockUser);
            setTokens(mockTokens);
            return;
        }

        const authUrl = await buildAuthorizationURL();
        window.location.href = authUrl;
    };

    // Initialize auth state from stored tokens
    useEffect(() => {
        const initAuth = async () => {
            const storedTokens = getStoredTokens();
            if (!storedTokens) {
                setIsLoading(false);
                return;
            }

            // Check if token is expired
            if (isTokenExpired(storedTokens)) {
                // Try to refresh
                if (storedTokens.refreshToken) {
                    try {
                        const newTokens = await refreshAccessToken(storedTokens.refreshToken);
                        storeTokens(newTokens);
                        setTokens(newTokens);
                        extractUser(newTokens);
                    } catch (error) {
                        console.error('Failed to refresh token:', error);
                        // Clear invalid tokens
                        localStorage.removeItem('auth_tokens');
                    }
                } else {
                    // No refresh token, clear tokens
                    localStorage.removeItem('auth_tokens');
                }
            } else {
                // Token is still valid
                setTokens(storedTokens);
                extractUser(storedTokens);
            }

            setIsLoading(false);
        };

        initAuth();
    }, []);

    // Auto-refresh token before expiration
    useEffect(() => {
        if (!tokens || !tokens.refreshToken) return;

        const checkAndRefresh = async () => {
            if (isTokenExpired(tokens)) {
                try {
                    const newTokens = await refreshAccessToken(tokens.refreshToken!);
                    storeTokens(newTokens);
                    setTokens(newTokens);
                    extractUser(newTokens);
                } catch (error) {
                    console.error('Failed to refresh token:', error);
                    logout();
                }
            }
        };

        // Check every minute
        const interval = setInterval(checkAndRefresh, 60 * 1000);

        return () => clearInterval(interval);
    }, [tokens]);

    const handleCallback = async (code: string, state: string) => {
        try {
            const newTokens = await exchangeCodeForTokens(code, state);
            storeTokens(newTokens);
            setTokens(newTokens);
            extractUser(newTokens);
        } catch (error) {
            console.error('Failed to exchange code for tokens:', error);
            throw error;
        }
    };

    const value: AuthContextType = {
        user,
        tokens,
        isAuthenticated: !!user && !!tokens,
        isLoading,
        login,
        logout,
        handleCallback,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
