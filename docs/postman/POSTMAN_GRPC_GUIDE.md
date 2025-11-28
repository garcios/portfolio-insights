# Postman gRPC Testing Guide

This guide explains how to test the gRPC services using Postman's native gRPC client.

## Prerequisites

- Postman v9.0 or later (supports gRPC)
- Services running locally (`make podman-up`)
- `.proto` files available in the `proto/` directory

## Quick Connection Details

| Service | Port | Proto File |
|---------|------|------------|
| **User Service** | `50051` | `proto/user/user.proto` |
| **Portfolio Service** | `50052` | `proto/portfolio/portfolio.proto` |
| **Transaction Service** | `50053` | `proto/transaction/transaction.proto` |
| **MarketData Service** | `50054` | `proto/marketdata/marketdata.proto` |

---

## Step-by-Step Guide

### 1. Create a New Request
1. Click **New** (top left) in Postman.
2. Select **gRPC Request**.

### 2. Configure Connection
1. **Enter URL**: `localhost:50051` (for User Service).
   - ⚠️ Do **NOT** use `http://` prefix.
2. **Security**: Click the **Padlock Icon** next to the URL and ensure it is **Unlocked** (Plaintext).
   - Our local dev environment does not use TLS.

### 3. Import Service Definition
1. Click the **Service Definition** tab.
2. Click **Import .proto file**.
3. Select the file: `proto/user/user.proto`.
   - If prompted for "Import Paths", add the root `proto/` directory.
4. The service `user.UserService` should now be loaded.

### 4. Invoke a Method
1. **Select Method**: Click the dropdown and choose `GetUser`.
2. **Enter Message**: In the **Message** tab, paste the JSON payload:
   ```json
   {
       "id": "YOUR_USER_UUID_HERE"
   }
   ```
   *(Example ID: `cc4c8131-a0a4-4d28-a482-35fefae52969`)*
3. Click **Invoke**.

### 5. Create a User (Example)
1. Change Method to `CreateUser`.
2. Payload:
   ```json
   {
       "email": "test@example.com",
       "name": "Test User",
       "password": "password123"
   }
   ```
3. Click **Invoke**.
4. Copy the returned `id` to use in `GetUser`.

## Troubleshooting

- **"Unavailable"**: Check if the service is running (`podman ps`) and the port matches the table above.
- **"Unimplemented"**: Ensure you selected the correct method and imported the correct `.proto` file.
- **"Internal"**: Check the service logs (`podman logs docker-compose_user-service_1`) for errors (e.g., DB connection issues).
