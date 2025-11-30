# How to Run the Portfolio Insights Application with OAuth2

This guide explains how to start the entire stack, including the new OAuth2/OIDC infrastructure.

## Prerequisites

- Docker or Podman installed
- Go 1.21+ installed
- Node.js 18+ installed

## Step 1: Start Infrastructure Services

Start the core services (PostgreSQL, NATS) and the new Hydra services (Hydra, Login Provider).

```bash
# Start core services (if not already running)
make podman-up

# Start Hydra services
make hydra-up
```

## Step 2: Configure OAuth Client

Run the script to register the frontend application as an OAuth2 client in Hydra.

```bash
./scripts/create-oauth-client.sh
```

## Step 3: Start Backend Services

You can run the backend services using `make` or individually.

```bash
# In a new terminal
make run-gateway
```

*Note: Ensure the Gateway is configured to use Hydra. Set `HYDRA_PUBLIC_URL=http://localhost:4444` in your environment or `.env` file.*

## Step 4: Start Frontend Application

```bash
cd apps/frontend

# Create .env.local from example
cp .env.example .env.local

# Install dependencies
npm install

# Start development server
npm run dev
```

## Step 5: Verify the Setup

1.  Open your browser to `http://localhost:5173`.
2.  Click the **"Sign In with OAuth"** button.
3.  You should be redirected to the Hydra login page (served by the Login Provider on port 3001).
4.  Enter any email and password (since it's a dev setup, it might accept any, or check `apps/login-consent-provider/handlers.go` for logic). *Actually, the current implementation checks against the database, so you might need to create a user first or use the registration flow if implemented.*
5.  Grant consent for the requested scopes.
6.  You should be redirected back to the app and logged in.

## Troubleshooting

-   **Hydra Connection Refused:** Ensure `podman-compose -f deployments/docker-compose/docker-compose.hydra.yml ps` shows all services running.
-   **Gateway Auth Error:** Check the Gateway logs. If `HYDRA_PUBLIC_URL` is not set, JWT validation might be disabled or failing.
-   **Frontend Redirect Loop:** Check the browser console and network tab. Ensure `VITE_REDIRECT_URI` matches exactly what was registered in `create-oauth-client.sh`.

## Useful Commands

-   **View Hydra Logs:** `podman logs -f hydra`
-   **View Login Provider Logs:** `podman logs -f login-consent-provider`
-   **Restart Hydra:** `podman-compose -f deployments/docker-compose/docker-compose.hydra.yml restart`
