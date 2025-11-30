#!/bin/bash

# Script to create OAuth2 client for Portfolio Insights SPA
# This should be run after Hydra is up and running

set -e

echo "Creating OAuth2 client for Portfolio Insights SPA..."

# Wait for Hydra to be ready
echo "Waiting for Hydra Admin API to be ready..."
until curl -s http://localhost:4445/health/ready > /dev/null; do
  echo "Hydra not ready yet, waiting..."
  sleep 2
done

echo "Hydra is ready!"

# Create the OAuth2 client
# Create the OAuth2 client using curl to specify a custom ID
curl -X POST http://localhost:4445/admin/clients \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "portfolio-insights-spa",
    "client_name": "Portfolio Insights SPA",
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code", "id_token"],
    "scope": "openid offline profile email portfolio:read portfolio:write transactions:read transactions:write",
    "redirect_uris": ["http://localhost:5173/auth/callback", "http://localhost:5173/"],
    "post_logout_redirect_uris": ["http://localhost:5173/"],
    "token_endpoint_auth_method": "none",
    "skip_consent": true
  }'

echo ""
echo "✅ OAuth2 client created successfully!"
echo ""
echo "Client ID: portfolio-insights-spa"
echo "Client Secret: (none - public client)"
echo "Redirect URIs: http://localhost:5173/auth/callback, http://localhost:5173/"
echo "Scopes: openid, offline, profile, email, portfolio:read, portfolio:write, transactions:read, transactions:write"
echo ""
echo "You can now use this client in your frontend application."
