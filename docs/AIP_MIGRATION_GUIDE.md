# AIP Compliance Migration Guide

## Overview

This guide provides step-by-step instructions for migrating from the old API to the new AIP-compliant API. The migration involves updating server handlers to use resource names and updating client code to construct proper resource name paths.

---

## What Changed?

### Key Changes Across All Services

1. **Resource Names**: All resources now use hierarchical resource names instead of simple IDs
   - Old: `id: "123"`
   - New: `name: "users/123"`

2. **Request Structures**:
   - Get/Delete: Use `name` field instead of `id`
   - List: Use `parent` field instead of `user_id`
   - Create: Accept resource object + `parent` field
   - Update: Accept resource object + `update_mask` field

3. **Response Changes**:
   - Delete now returns `google.protobuf.Empty` instead of custom response
   - All resources include both `name` and extracted ID fields

4. **HTTP Annotations**: All RPCs now have HTTP/REST mappings

---

## Service-by-Service Migration

### 1. User Service

#### Proto Changes

**Before:**
```protobuf
message GetUserRequest {
  string id = 1;
}

message CreateUserRequest {
  string email = 1;
  string username = 2;
  string password = 3;
}
```

**After:**
```protobuf
message User {
  string name = 1;  // Format: users/{user}
  string email = 2;
  string username = 3;
  string password = 4;  // INPUT_ONLY
  string user_id = 10;  // OUTPUT_ONLY
}

message GetUserRequest {
  string name = 1;  // Format: users/{user}
}

message CreateUserRequest {
  User user = 1;
  string user_id = 2;  // Optional
}
```

#### Server Handler Migration

**Before:**
```go
func (s *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
    userID := req.Id
    
    user, err := s.repo.FindByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    return &userpb.GetUserResponse{
        Id:       user.ID,
        Email:    user.Email,
        Username: user.Username,
    }, nil
}
```

**After:**
```go
import "github.com/garcios/portfolio-insights/pkg/resourcenames"

func (s *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.User, error) {
    // Parse resource name
    userID, err := resourcenames.ParseUserName(req.Name)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
    }
    
    user, err := s.repo.FindByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    return &userpb.User{
        Name:     resourcenames.UserName(user.ID),
        Email:    user.Email,
        Username: user.Username,
        UserId:   user.ID,
    }, nil
}
```

**CreateUser - Before:**
```go
func (s *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
    user := &domain.User{
        Email:    req.Email,
        Username: req.Username,
        Password: req.Password,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return &userpb.CreateUserResponse{
        Id: user.ID,
    }, nil
}
```

**CreateUser - After:**
```go
func (s *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.User, error) {
    // Extract user data from request
    user := &domain.User{
        Email:    req.User.Email,
        Username: req.User.Username,
        Password: req.User.Password,
    }
    
    // Use client-specified ID if provided
    if req.UserId != "" {
        user.ID = req.UserId
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return &userpb.User{
        Name:     resourcenames.UserName(user.ID),
        Email:    user.Email,
        Username: user.Username,
        UserId:   user.ID,
    }, nil
}
```

#### Client Code Migration

**Before:**
```go
resp, err := client.GetUser(ctx, &userpb.GetUserRequest{
    Id: "123",
})
```

**After:**
```go
import "github.com/garcios/portfolio-insights/pkg/resourcenames"

resp, err := client.GetUser(ctx, &userpb.GetUserRequest{
    Name: resourcenames.UserName("123"),
})
```

---

### 2. Transaction Service

#### Proto Changes

**Before:**
```protobuf
message GetTransactionRequest {
  string id = 1;
}

message ListTransactionsRequest {
  string user_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message UpdateTransactionRequest {
  string id = 1;
  string type = 3;
  optional string symbol = 2;
  // ... many individual fields
}

message DeleteTransactionResponse {
  bool success = 1;
}
```

**After:**
```protobuf
message Transaction {
  string name = 1;  // Format: users/{user}/transactions/{transaction}
  string user_id = 2;  // OUTPUT_ONLY
  // ... other fields
  string transaction_id = 20;  // OUTPUT_ONLY
}

message GetTransactionRequest {
  string name = 1;  // Format: users/{user}/transactions/{transaction}
}

message ListTransactionsRequest {
  string parent = 1;  // Format: users/{user}
  int32 page_size = 2;
  string page_token = 3;
  TransactionFilter filter = 4;
}

message UpdateTransactionRequest {
  Transaction transaction = 1;
  google.protobuf.FieldMask update_mask = 2;
}

// DeleteTransaction now returns google.protobuf.Empty
```

#### Server Handler Migration

**GetTransaction - Before:**
```go
func (h *TransactionHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
    txn, err := h.repo.FindByID(ctx, req.Id)
    if err != nil {
        return nil, err
    }
    
    return &pb.GetTransactionResponse{
        Transaction: convertToProto(txn),
    }, nil
}
```

**GetTransaction - After:**
```go
import "github.com/garcios/portfolio-insights/pkg/resourcenames"

func (h *TransactionHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.Transaction, error) {
    // Parse hierarchical resource name
    userID, txnID, err := resourcenames.ParseTransactionName(req.Name)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
    }
    
    // Verify transaction belongs to user (security check)
    txn, err := h.repo.FindByUserAndID(ctx, userID, txnID)
    if err != nil {
        return nil, err
    }
    
    return convertToProtoWithName(txn), nil
}

func convertToProtoWithName(txn *domain.Transaction) *pb.Transaction {
    return &pb.Transaction{
        Name:          resourcenames.TransactionName(txn.UserID, txn.ID),
        UserId:        txn.UserID,
        Type:          txn.Type,
        Symbol:        &txn.Symbol,
        // ... other fields
        TransactionId: txn.ID,
    }
}
```

**ListTransactions - Before:**
```go
func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
    txns, nextToken, err := h.repo.List(ctx, req.UserId, req.PageSize, req.PageToken)
    if err != nil {
        return nil, err
    }
    
    return &pb.ListTransactionsResponse{
        Transactions:  convertManyToProto(txns),
        NextPageToken: nextToken,
    }, nil
}
```

**ListTransactions - After:**
```go
func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
    // Parse parent resource name
    userID, err := resourcenames.ParseTransactionParent(req.Parent)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid parent: %v", err)
    }
    
    txns, nextToken, err := h.repo.List(ctx, userID, req.PageSize, req.PageToken)
    if err != nil {
        return nil, err
    }
    
    return &pb.ListTransactionsResponse{
        Transactions:  convertManyToProtoWithNames(txns),
        NextPageToken: nextToken,
    }, nil
}
```

**UpdateTransaction - Before:**
```go
func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
    txn := &domain.Transaction{
        ID:           req.Id,
        Type:         req.Type,
        Symbol:       *req.Symbol,
        Quantity:     *req.Quantity,
        // ... many fields
    }
    
    if err := h.repo.Update(ctx, txn); err != nil {
        return nil, err
    }
    
    return &pb.UpdateTransactionResponse{
        Transaction: convertToProto(txn),
    }, nil
}
```

**UpdateTransaction - After:**
```go
import "google.golang.org/protobuf/types/known/fieldmaskpb"

func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.Transaction, error) {
    // Parse resource name
    userID, txnID, err := resourcenames.ParseTransactionName(req.Transaction.Name)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
    }
    
    // Get existing transaction
    existing, err := h.repo.FindByUserAndID(ctx, userID, txnID)
    if err != nil {
        return nil, err
    }
    
    // Apply field mask
    if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
        // Update all fields
        applyAllUpdates(existing, req.Transaction)
    } else {
        // Update only specified fields
        for _, path := range req.UpdateMask.Paths {
            applyFieldUpdate(existing, req.Transaction, path)
        }
    }
    
    if err := h.repo.Update(ctx, existing); err != nil {
        return nil, err
    }
    
    return convertToProtoWithName(existing), nil
}

func applyFieldUpdate(existing *domain.Transaction, update *pb.Transaction, path string) {
    switch path {
    case "type":
        existing.Type = update.Type
    case "symbol":
        if update.Symbol != nil {
            existing.Symbol = *update.Symbol
        }
    case "quantity":
        if update.Quantity != nil {
            existing.Quantity = *update.Quantity
        }
    // ... handle other fields
    }
}
```

**DeleteTransaction - Before:**
```go
func (h *TransactionHandler) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
    if err := h.repo.Delete(ctx, req.Id); err != nil {
        return nil, err
    }
    
    return &pb.DeleteTransactionResponse{
        Success: true,
    }, nil
}
```

**DeleteTransaction - After:**
```go
import "google.golang.org/protobuf/types/known/emptypb"

func (h *TransactionHandler) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*emptypb.Empty, error) {
    // Parse resource name
    userID, txnID, err := resourcenames.ParseTransactionName(req.Name)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
    }
    
    // Verify ownership before deleting
    if err := h.repo.DeleteByUserAndID(ctx, userID, txnID); err != nil {
        return nil, err
    }
    
    return &emptypb.Empty{}, nil
}
```

#### Client Code Migration

**Before:**
```go
// Get
resp, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{
    Id: "txn-123",
})

// List
resp, err := client.ListTransactions(ctx, &pb.ListTransactionsRequest{
    UserId:   "user-456",
    PageSize: 50,
})

// Update
resp, err := client.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
    Id:     "txn-123",
    Type:   "SELL",
    Symbol: proto.String("AAPL"),
})

// Delete
resp, err := client.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
    Id: "txn-123",
})
```

**After:**
```go
import (
    "github.com/garcios/portfolio-insights/pkg/resourcenames"
    "google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Get
resp, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{
    Name: resourcenames.TransactionName("user-456", "txn-123"),
})

// List
resp, err := client.ListTransactions(ctx, &pb.ListTransactionsRequest{
    Parent:   resourcenames.UserName("user-456"),
    PageSize: 50,
})

// Update
resp, err := client.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
    Transaction: &pb.Transaction{
        Name:   resourcenames.TransactionName("user-456", "txn-123"),
        Type:   "SELL",
        Symbol: proto.String("AAPL"),
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"type", "symbol"},
    },
})

// Delete
_, err := client.DeleteTransaction(ctx, &pb.DeleteTransactionRequest{
    Name: resourcenames.TransactionName("user-456", "txn-123"),
})
```

---

### 3. Portfolio Service

#### Key Changes

- Portfolio is now a singleton resource: `users/{user}/portfolio`
- Holdings use hierarchical names: `users/{user}/holdings/{holding}`
- All requests use `name` or `parent` fields

#### Server Handler Migration

**Before:**
```go
func (h *PortfolioHandler) GetPortfolioSummary(ctx context.Context, req *pb.GetPortfolioSummaryRequest) (*pb.GetPortfolioSummaryResponse, error) {
    summary, err := h.usecase.GetSummary(ctx, req.UserId)
    if err != nil {
        return nil, err
    }
    
    return &pb.GetPortfolioSummaryResponse{
        Summary: summary,
    }, nil
}
```

**After:**
```go
func (h *PortfolioHandler) GetPortfolioSummary(ctx context.Context, req *pb.GetPortfolioSummaryRequest) (*pb.PortfolioSummary, error) {
    // Parse portfolio resource name (singleton)
    userID, err := resourcenames.ParsePortfolioName(req.Name)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
    }
    
    summary, err := h.usecase.GetSummary(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Add resource name to response
    summary.Name = resourcenames.PortfolioName(userID)
    summary.UserId = userID
    
    return summary, nil
}
```

---

### 4. Market Data Service

#### Key Changes

- Assets use resource names: `assets/{asset}`
- Added `GetAssetBySymbol` custom method for symbol-based lookup
- Price methods now accept resource names

#### Server Handler Migration

**Before:**
```go
func (h *MarketDataHandler) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
    asset, err := h.repo.FindBySymbol(ctx, req.Symbol)
    if err != nil {
        return nil, err
    }
    
    return &pb.GetAssetResponse{
        Asset: convertToProto(asset),
    }, nil
}
```

**After:**
```go
func (h *MarketDataHandler) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.Asset, error) {
    // Parse asset resource name
    assetID, err := resourcenames.ParseAssetName(req.Name)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid resource name: %v", err)
    }
    
    asset, err := h.repo.FindByID(ctx, assetID)
    if err != nil {
        return nil, err
    }
    
    return &pb.Asset{
        Name:        resourcenames.AssetName(asset.ID),
        Symbol:      asset.Symbol,
        DisplayName: asset.Name,
        Type:        asset.Type,
        Exchange:    asset.Exchange,
        Currency:    asset.Currency,
        AssetId:     asset.ID,
    }, nil
}

// New custom method for symbol-based lookup
func (h *MarketDataHandler) GetAssetBySymbol(ctx context.Context, req *pb.GetAssetBySymbolRequest) (*pb.Asset, error) {
    asset, err := h.repo.FindBySymbol(ctx, req.Symbol)
    if err != nil {
        return nil, err
    }
    
    return &pb.Asset{
        Name:        resourcenames.AssetName(asset.ID),
        Symbol:      asset.Symbol,
        DisplayName: asset.Name,
        Type:        asset.Type,
        Exchange:    asset.Exchange,
        Currency:    asset.Currency,
        AssetId:     asset.ID,
    }, nil
}
```

---

## Migration Steps

### Phase 1: Preparation (No Downtime)

1. **Review Changes**
   - Read this migration guide
   - Review updated proto files
   - Understand resource name patterns

2. **Update Dependencies**
   - Ensure `make proto-gen` runs successfully
   - Verify Google proto dependencies are downloaded

3. **Test Resource Name Helpers**
   ```bash
   cd pkg/resourcenames && go test -v
   ```

### Phase 2: Server Updates (Requires Deployment)

1. **Update User Service**
   - Update `services/user-service/internal/handler/user_handler.go`
   - Run tests: `make test-user`
   - Build: `make build-user-service`

2. **Update Transaction Service**
   - Update `services/transaction-service/internal/handler/transaction_handler.go`
   - Update repository methods if needed (add `FindByUserAndID`, `DeleteByUserAndID`)
   - Run tests: `make test-transaction`
   - Build: `make build-transaction-service`

3. **Update Portfolio Service**
   - Update `services/portfolio-service/internal/handler/portfolio_handler.go`
   - Run tests: `make test-portfolio`
   - Build: `make build-portfolio-service`

4. **Update Market Data Service**
   - Update `services/marketdata-service/internal/handler/marketdata_handler.go`
   - Add `GetAssetBySymbol` handler
   - Run tests: `make test-marketdata`
   - Build: `make build-marketdata-service`

5. **Update Gateway**
   - Update GraphQL resolvers to use new resource names
   - Update any direct gRPC calls
   - Run tests: `make test-gateway`

### Phase 3: Deployment

1. **Deploy Services**
   ```bash
   make services-down
   make services-up
   ```

2. **Verify Services**
   - Check logs: `make services-logs`
   - Test each endpoint manually
   - Run integration tests

### Phase 4: Client Updates

1. **Update Client Code**
   - Import `pkg/resourcenames`
   - Update all gRPC calls to use resource names
   - Update field mask usage for updates

2. **Test Client Code**
   - Unit tests
   - Integration tests
   - Manual testing

---

## Testing Strategy

### Unit Tests

Update unit tests to use resource names:

```go
// Before
req := &pb.GetUserRequest{
    Id: "test-user-123",
}

// After
req := &pb.GetUserRequest{
    Name: resourcenames.UserName("test-user-123"),
}
```

### Integration Tests

Update integration test scripts:

```bash
# Before
grpcurl -d '{"id": "123"}' localhost:50051 user.UserService/GetUser

# After
grpcurl -d '{"name": "users/123"}' localhost:50051 user.UserService/GetUser
```

### Manual Testing

Use grpcurl or Postman to test:

```bash
# Get User
grpcurl -d '{"name": "users/123"}' localhost:50051 user.UserService/GetUser

# List Transactions
grpcurl -d '{"parent": "users/123", "page_size": 10}' localhost:50052 transaction.TransactionService/ListTransactions

# Update Transaction with Field Mask
grpcurl -d '{
  "transaction": {
    "name": "users/123/transactions/txn-456",
    "type": "SELL"
  },
  "update_mask": {
    "paths": ["type"]
  }
}' localhost:50052 transaction.TransactionService/UpdateTransaction
```

---

## Rollback Procedure

If issues arise during migration:

1. **Revert Proto Files**
   ```bash
   git checkout HEAD~1 proto/
   make proto-gen
   ```

2. **Revert Handler Changes**
   ```bash
   git checkout HEAD~1 services/*/internal/handler/
   ```

3. **Rebuild and Redeploy**
   ```bash
   make services-down
   make services-up
   ```

---

## Common Issues and Solutions

### Issue 1: "Invalid resource name" errors

**Cause**: Resource name format doesn't match expected pattern

**Solution**: Use resource name helpers:
```go
name := resourcenames.TransactionName(userID, txnID)
```

### Issue 2: Field mask not working

**Cause**: Field mask paths don't match proto field names

**Solution**: Use exact proto field names (snake_case):
```go
UpdateMask: &fieldmaskpb.FieldMask{
    Paths: []string{"price_per_share", "quantity"},  // Not "pricePerShare"
}
```

### Issue 3: Empty response from Delete

**Cause**: Expecting old `DeleteTransactionResponse`

**Solution**: Delete now returns `Empty`:
```go
_, err := client.DeleteTransaction(ctx, req)
// No response body to check
```

### Issue 4: Missing user ID in hierarchical names

**Cause**: Trying to access transaction without user context

**Solution**: Always include user ID in transaction names:
```go
// Wrong
name := "transactions/txn-123"

// Correct
name := resourcenames.TransactionName("user-456", "txn-123")
```

---

## Benefits After Migration

1. **Consistency**: All APIs follow Google's AIP standards
2. **Security**: Hierarchical names enforce parent-child relationships
3. **Clarity**: Resource paths are self-documenting
4. **Tooling**: Better support for code generation and client libraries
5. **HTTP/REST**: Proper REST mappings via gRPC-HTTP transcoding
6. **Field Masks**: Efficient partial updates

---

## Next Steps

1. Complete server-side migration
2. Update all client code
3. Update integration tests
4. Update API documentation
5. Consider implementing gRPC-HTTP transcoding for REST access

---

## Support

For questions or issues during migration:
1. Review this guide
2. Check the AIP compliance evaluation document
3. Review test examples in `pkg/resourcenames/resourcenames_test.go`
4. Consult Google AIP documentation: https://google.aip.dev/
