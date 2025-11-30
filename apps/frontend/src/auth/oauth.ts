/**
 * OAuth2/OIDC utilities for Ory Hydra integration
 */

const HYDRA_AUTH_URL = import.meta.env.VITE_HYDRA_AUTH_URL || 'http://localhost:4444/oauth2/auth';
const HYDRA_TOKEN_URL = import.meta.env.VITE_HYDRA_TOKEN_URL || 'http://localhost:4444/oauth2/token';
const HYDRA_LOGOUT_URL = import.meta.env.VITE_HYDRA_LOGOUT_URL || 'http://localhost:4444/oauth2/sessions/logout';
const CLIENT_ID = import.meta.env.VITE_CLIENT_ID || 'portfolio-insights-spa';
const REDIRECT_URI = import.meta.env.VITE_REDIRECT_URI || 'http://localhost:5173/auth/callback';

export interface TokenResponse {
    access_token: string;
    token_type: string;
    expires_in: number;
    refresh_token?: string;
    id_token?: string;
    scope: string;
}

export interface AuthTokens {
    accessToken: string;
    refreshToken?: string;
    idToken?: string;
    expiresAt: number;
}

/**
 * Generate a random string for PKCE code verifier
 */
function generateRandomString(length: number): string {
    const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
    const randomValues = new Uint8Array(length);
    crypto.getRandomValues(randomValues);
    return Array.from(randomValues)
        .map(v => charset[v % charset.length])
        .join('');
}

/**
 * Generate SHA-256 hash and base64url encode
 */
async function sha256(plain: string): Promise<string> {
    const encoder = new TextEncoder();
    const data = encoder.encode(plain);
    const hash = await crypto.subtle.digest('SHA-256', data);
    return base64urlEncode(hash);
}

/**
 * Base64URL encode
 */
function base64urlEncode(buffer: ArrayBuffer): string {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary)
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
}

/**
 * Generate PKCE code verifier and challenge
 */
export async function generatePKCE(): Promise<{ verifier: string; challenge: string }> {
    const verifier = generateRandomString(128);
    const challenge = await sha256(verifier);
    return { verifier, challenge };
}

/**
 * Build authorization URL with PKCE
 */
export async function buildAuthorizationURL(scopes: string[] = ['openid', 'offline', 'profile', 'email']): Promise<string> {
    const { verifier, challenge } = await generatePKCE();
    const state = generateRandomString(32);

    // Store verifier and state in sessionStorage
    sessionStorage.setItem('pkce_verifier', verifier);
    sessionStorage.setItem('oauth_state', state);

    const params = new URLSearchParams({
        client_id: CLIENT_ID,
        response_type: 'code',
        redirect_uri: REDIRECT_URI,
        scope: scopes.join(' '),
        state,
        code_challenge: challenge,
        code_challenge_method: 'S256',
    });

    return `${HYDRA_AUTH_URL}?${params.toString()}`;
}

/**
 * Exchange authorization code for tokens
 */
export async function exchangeCodeForTokens(code: string, state: string): Promise<AuthTokens> {
    // Verify state
    const storedState = sessionStorage.getItem('oauth_state');
    if (state !== storedState) {
        throw new Error('Invalid state parameter');
    }

    // Get code verifier
    const verifier = sessionStorage.getItem('pkce_verifier');
    if (!verifier) {
        throw new Error('Missing PKCE verifier');
    }

    // Exchange code for tokens
    const params = new URLSearchParams({
        grant_type: 'authorization_code',
        code,
        redirect_uri: REDIRECT_URI,
        client_id: CLIENT_ID,
        code_verifier: verifier,
    });

    const response = await fetch(HYDRA_TOKEN_URL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: params.toString(),
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Token exchange failed: ${error}`);
    }

    const tokenResponse: TokenResponse = await response.json();

    // Clean up session storage
    sessionStorage.removeItem('pkce_verifier');
    sessionStorage.removeItem('oauth_state');

    // Calculate expiration time
    const expiresAt = Date.now() + tokenResponse.expires_in * 1000;

    return {
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token,
        idToken: tokenResponse.id_token,
        expiresAt,
    };
}

/**
 * Refresh access token using refresh token
 */
export async function refreshAccessToken(refreshToken: string): Promise<AuthTokens> {
    const params = new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: CLIENT_ID,
    });

    const response = await fetch(HYDRA_TOKEN_URL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: params.toString(),
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Token refresh failed: ${error}`);
    }

    const tokenResponse: TokenResponse = await response.json();

    // Calculate expiration time
    const expiresAt = Date.now() + tokenResponse.expires_in * 1000;

    return {
        accessToken: tokenResponse.access_token,
        refreshToken: tokenResponse.refresh_token || refreshToken, // Use new refresh token if provided
        idToken: tokenResponse.id_token,
        expiresAt,
    };
}

/**
 * Initiate logout
 */
export function logout(): void {
    // Clear tokens from storage
    localStorage.removeItem('auth_tokens');
    sessionStorage.clear();

    // Redirect to Hydra logout
    window.location.href = `${HYDRA_LOGOUT_URL}?post_logout_redirect_uri=${encodeURIComponent(window.location.origin)}`;
}

/**
 * Store tokens in localStorage
 */
export function storeTokens(tokens: AuthTokens): void {
    localStorage.setItem('auth_tokens', JSON.stringify(tokens));
}

/**
 * Get tokens from localStorage
 */
export function getStoredTokens(): AuthTokens | null {
    const stored = localStorage.getItem('auth_tokens');
    if (!stored) return null;

    try {
        return JSON.parse(stored);
    } catch {
        return null;
    }
}

/**
 * Check if access token is expired or about to expire (within 5 minutes)
 */
export function isTokenExpired(tokens: AuthTokens): boolean {
    const fiveMinutes = 5 * 60 * 1000;
    return tokens.expiresAt - Date.now() < fiveMinutes;
}

/**
 * Decode JWT token (without verification - for display purposes only)
 */
export function decodeJWT(token: string): any {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) return null;

        const payload = parts[1];
        const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
        return JSON.parse(decoded);
    } catch {
        return null;
    }
}
