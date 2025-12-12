# User Service Test Client

A gRPC test client for the User Service that allows you to test all available RPC methods.

## Prerequisites

- Go 1.24.0 or later
- User Service running (default: `localhost:50051`)

## Installation

```bash
cd apps/test-client/user-service
go mod download
go build -o user-client
```

## Usage

The test client supports four operations: `get`, `create`, `verify`, and `test-errors`.

### Get User

Retrieve a user by their ID:

```bash
./user-client -op get -user-id 123
```

With custom server address:

```bash
./user-client -addr localhost:50051 -op get -user-id 123
```

### Create User

Create a new user:

```bash
./user-client -op create -email john@example.com -username johndoe -password secretpass
```

Create a user with a specific ID:

```bash
./user-client -op create -email john@example.com -username johndoe -password secretpass -user-id custom-id-123
```

### Verify User

Verify user credentials:

```bash
./user-client -op verify -email john@example.com -password secretpass
```

### Test Error Handling

Run a comprehensive suite of error tests to validate input validation and error responses:

```bash
./user-client -op test-errors
```

This will run 10 different error test cases including:
- Empty and invalid resource names
- Missing required fields
- Invalid email formats
- Non-existent users
- Invalid UUID formats

Use the `-verbose` flag to see additional error details:

```bash
./user-client -op test-errors -verbose
```

## Command-Line Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `-addr` | Server address (host:port) | `localhost:50051` | No |
| `-op` | Operation to perform (get, create, verify, test-errors) | `get` | No |
| `-user-id` | User ID for get operation | - | Yes (for get) |
| `-email` | Email address | - | Yes (for create/verify) |
| `-username` | Username | - | Yes (for create) |
| `-password` | Password | - | Yes (for create/verify) |
| `-verbose` | Enable verbose error output | `false` | No |

## Examples

### Complete Workflow

1. **Create a new user:**
   ```bash
   ./user-client -op create -email alice@example.com -username alice -password mypassword
   ```
   
   Output:
   ```
   Creating user: email=alice@example.com, username=alice
   User created successfully!
   
   === User Details ===
   Resource Name: users/1
   User ID:       1
   Email:         alice@example.com
   Username:      alice
   ===================
   ```

2. **Retrieve the user:**
   ```bash
   ./user-client -op get -user-id 1
   ```

3. **Verify credentials:**
   ```bash
   ./user-client -op verify -email alice@example.com -password mypassword
   ```
   
   Output:
   ```
   Verifying user: email=alice@example.com
   ✓ Credentials are VALID
   
   === User Details ===
   Resource Name: users/1
   User ID:       1
   Email:         alice@example.com
   Username:      alice
   ===================
   ```

### Testing Against Different Environments

**Local development:**
```bash
./user-client -addr localhost:50051 -op get -user-id 1
```

**Docker environment:**
```bash
./user-client -addr user-service:50051 -op get -user-id 1
```

## Error Handling

The client provides clear error messages for common issues:

- **Connection failures:** Check if the server is running and the address is correct
- **Missing required fields:** The client will indicate which fields are required for each operation
- **Invalid credentials:** Verify operations will clearly indicate if credentials are valid or invalid
- **User not found:** Get operations will return an error if the user doesn't exist

## Development

To run without building:

```bash
go run main.go -op get -user-id 123
```

To build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o user-client-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o user-client-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o user-client.exe
```
