# Transaction Service Test Client

A gRPC test client for the Transaction Service that allows you to test all available RPC methods and validate error handling.

## Prerequisites

- Go 1.24.0 or later
- Transaction Service running (default: `localhost:50053`)

## Installation

```bash
cd apps/test-client/transaction-service
go mod download
go build -o transaction-client
```

## Usage

The test client supports seven operations: `create`, `get`, `list`, `update`, `delete`, `oldest`, and `test-errors`.

### Create Transaction

Create transactions for different types: equity (BUY, SELL) and cash (DEP, WIT, INT, DIV).

**Create a BUY transaction:**
```bash
./transaction-client -op create -user-id user-123 \
  -type BUY -symbol AAPL -quantity 10 -price 150.50 \
  -executed-at "2024-01-15T10:30:00Z" -brokerage 5.00
```

**Create a deposit (DEP) transaction:**
```bash
./transaction-client -op create -user-id user-123 \
  -type DEP -amount 5000.00 -executed-at "2024-01-01T09:00:00Z"
```

**Create a dividend (DIV) transaction:**
```bash
./transaction-client -op create -user-id user-123 \
  -type DIV -symbol AAPL -amount 25.50 -executed-at "2024-03-15T12:00:00Z"
```

### Get Transaction

Retrieve a specific transaction by its ID:

```bash
./transaction-client -op get -user-id user-123 -transaction-id txn-456
```

### List Transactions

List all transactions for a user with optional filtering:

```bash
# List all transactions
./transaction-client -op list -user-id user-123

# List with pagination
./transaction-client -op list -user-id user-123 -page-size 10

# Filter by symbol
./transaction-client -op list -user-id user-123 -filter-symbol AAPL

# Filter by type
./transaction-client -op list -user-id user-123 -filter-type BUY
```

### Update Transaction

Update an existing transaction using field masks:

```bash
# Update notes only
./transaction-client -op update -user-id user-123 -transaction-id txn-456 \
  -notes "Updated notes" -update-fields notes

# Update multiple fields
./transaction-client -op update -user-id user-123 -transaction-id txn-456 \
  -brokerage 7.50 -notes "Revised transaction" \
  -update-fields brokerage,notes
```

### Delete Transaction

Delete a transaction:

```bash
./transaction-client -op delete -user-id user-123 -transaction-id txn-456
```

### Get Oldest Transaction

Retrieve the oldest transaction for a user:

```bash
./transaction-client -op oldest -user-id user-123
```

### Test Error Handling

Run a comprehensive suite of error tests to validate input validation and error responses:

```bash
./transaction-client -op test-errors
```

This will run 16 different error test cases including:
- Empty and invalid resource names
- Missing required fields (type, executed_at)
- Invalid parent formats
- Non-existent transactions
- Nil transaction objects
- Invalid field masks

Use the `-verbose` flag to see additional error details:

```bash
./transaction-client -op test-errors -verbose
```

## Command-Line Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `-addr` | Server address (host:port) | `localhost:50053` | No |
| `-op` | Operation to perform | `create` | No |
| `-user-id` | User ID for parent resource name | - | Yes (most ops) |
| `-transaction-id` | Transaction ID | - | Yes (get/update/delete) |
| `-type` | Transaction type (BUY, SELL, DEP, WIT, INT, DIV) | - | Yes (create) |
| `-symbol` | Asset symbol | - | Conditional* |
| `-quantity` | Quantity | `0` | Conditional* |
| `-price` | Price per share | `0` | Conditional* |
| `-amount` | Cash amount | `0` | Conditional* |
| `-executed-at` | Execution timestamp (RFC3339) | - | Yes (create) |
| `-brokerage` | Brokerage fee | `0` | No |
| `-notes` | Transaction notes | - | No |
| `-price-currency` | Price currency | `USD` | No |
| `-brokerage-currency` | Brokerage currency | `USD` | No |
| `-filter-symbol` | Filter by symbol (list) | - | No |
| `-filter-type` | Filter by type (list) | - | No |
| `-page-size` | Page size (list) | `50` | No |
| `-page-token` | Page token (list) | - | No |
| `-update-fields` | Field paths for update | - | No |
| `-verbose` | Enable verbose error output | `false` | No |

\* **Conditional requirements:**
- BUY/SELL transactions require: `symbol`, `quantity`, `price`
- DIV transactions require: `symbol`, `amount`
- DEP/WIT/INT transactions require: `amount`

## Transaction Types

### Equity Transactions

**BUY** - Purchase of securities
- Required: `symbol`, `quantity`, `price`
- Example: Buy 10 shares of AAPL at $150.50

**SELL** - Sale of securities
- Required: `symbol`, `quantity`, `price`
- Example: Sell 5 shares of AAPL at $155.75

### Cash Transactions

**DEP** - Deposit
- Required: `amount`
- Example: Deposit $5000 into account

**WIT** - Withdrawal
- Required: `amount`
- Example: Withdraw $1000 from account

**INT** - Interest
- Required: `amount`
- Example: Interest payment of $15.50

**DIV** - Dividend
- Required: `symbol`, `amount`
- Example: Dividend of $25.50 from AAPL

## Examples

### Complete Workflow

1. **Create a BUY transaction:**
   ```bash
   ./transaction-client -op create -user-id alice \
     -type BUY -symbol AAPL -quantity 10 -price 150.50 \
     -executed-at "2024-01-15T10:30:00Z" -brokerage 5.00
   ```
   
   Output:
   ```
   Creating BUY transaction for user alice
   ✓ Transaction created successfully!
   
   === Transaction Details ===
   Resource Name:     users/alice/transactions/abc-123
   Transaction ID:    abc-123
   User ID:           alice
   Type:              BUY
   Symbol:            AAPL
   Quantity:          10.0000
   Price per Share:   150.50 USD
   Executed At:       2024-01-15T10:30:00Z
   Brokerage:         5.00 USD
   ===========================
   ```

2. **List transactions:**
   ```bash
   ./transaction-client -op list -user-id alice -filter-symbol AAPL
   ```

3. **Update transaction:**
   ```bash
   ./transaction-client -op update -user-id alice -transaction-id abc-123 \
     -notes "Initial AAPL purchase" -update-fields notes
   ```

4. **Get oldest transaction:**
   ```bash
   ./transaction-client -op oldest -user-id alice
   ```

5. **Delete transaction:**
   ```bash
   ./transaction-client -op delete -user-id alice -transaction-id abc-123
   ```

### Testing Against Different Environments

**Local development:**
```bash
./transaction-client -addr localhost:50053 -op list -user-id alice
```

**Docker environment:**
```bash
./transaction-client -addr transaction-service:50053 -op list -user-id alice
```

## Error Handling

The client provides clear error messages for common issues:

- **Connection failures:** Check if the server is running and the address is correct
- **Missing required fields:** The client will indicate which fields are required for each operation
- **Invalid resource names:** Proper format is `users/{user}/transactions/{transaction}`
- **Transaction not found:** Get/update/delete operations will return NotFound for non-existent transactions
- **Invalid transaction types:** Must be one of: BUY, SELL, DEP, WIT, INT, DIV

## Development

To run without building:

```bash
go run main.go -op create -user-id test -type DEP -amount 1000 -executed-at "2024-01-01T00:00:00Z"
```

To build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o transaction-client-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o transaction-client-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o transaction-client.exe
```

## Scripts

### examples.sh

Demonstrates a complete workflow including:
- Creating various transaction types (BUY, SELL, DEP, DIV)
- Listing with filters
- Updating transactions
- Getting oldest transaction
- Deleting transactions

```bash
./examples.sh
```

### demo-errors.sh

Showcases error handling capabilities:
- Invalid inputs
- Missing required fields
- Non-existent resources
- Comprehensive error test suite

```bash
./demo-errors.sh
```
