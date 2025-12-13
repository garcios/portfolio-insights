# AIP Compliance Evaluation Report

## Executive Summary

This report evaluates the Protocol Buffer definitions in the portfolio-insights project against Google's API Improvement Proposals (AIP) standards. The evaluation covers four proto files: `marketdata.proto`, `portfolio.proto`, `transaction.proto`, and `user.proto`.

**Overall Assessment**: The proto files demonstrate good documentation practices and generally follow many AIP conventions. However, there are several significant deviations from AIP standards, particularly around resource naming, standard method patterns, and field behavior annotations.

---

## Evaluated Files

1. `proto/marketdata/marketdata.proto`
2. `proto/portfolio/portfolio.proto`
3. `proto/transaction/transaction.proto`
4. `proto/user/user.proto`

---

## Compliance Analysis by AIP Category

### ✅ **Compliant Areas**

#### AIP-191: File and Directory Structure
- **Status**: ✅ Fully Compliant
- All files use `proto3` syntax
- Files use `snake_case` naming
- Proper file layout with syntax, package, imports, and service definitions in correct order
- Good separation of concerns with one service per file

#### AIP-192: Documentation
- **Status**: ✅ Mostly Compliant
- All services, RPCs, messages, and fields have documentation comments
- Comments are clear and descriptive
- Good use of examples in field comments (e.g., "AAPL", "USD")

#### AIP-140: Field Names
- **Status**: ✅ Mostly Compliant
- Fields use `lower_snake_case` consistently
- Repeated fields use proper plural forms (`assets`, `transactions`, `holdings`)
- Singular fields use singular forms (`asset`, `transaction`)
- Good use of standard abbreviations (`id` instead of `identifier`)
- Boolean fields avoid "is_" prefix (e.g., `success`, `valid`)

#### AIP-142: Time and Duration
- **Status**: ✅ Compliant
- Proper use of `google.protobuf.Timestamp` for all time fields
- Consistent field naming (`created_at`, `updated_at`, `executed_at`)

---

### ⚠️ **Non-Compliant or Partially Compliant Areas**

#### AIP-122: Resource Names
- **Status**: ❌ **Major Non-Compliance**
- **Issue**: Resources do not follow the hierarchical resource name pattern
- **Current**: Resources use simple `id` fields (e.g., `string id = 1`)
- **Expected**: Resources should use `name` fields with hierarchical paths

**Examples of violations:**

```protobuf
// Current (Non-compliant)
message Asset {
  string id = 1;
  string symbol = 2;
  // ...
}

// AIP-122 Compliant
message Asset {
  // Resource name in format: assets/{asset}
  string name = 1;
  string symbol = 2;
  // Optional: string asset_id = 10 [(google.api.field_behavior) = OUTPUT_ONLY];
}
```

**Impact**: 
- All resource messages (Asset, Transaction, User, etc.) violate this standard
- Request messages use `id` or `symbol` instead of `name`
- No hierarchical resource structure (e.g., `users/{user}/transactions/{transaction}`)

---

#### AIP-131: Standard Methods - Get
- **Status**: ⚠️ **Partially Compliant**
- **Issues**:
  1. Request fields should be named `name` instead of `id` or `symbol`
  2. Missing HTTP annotations (`google.api.http`)
  3. Missing method signature annotations (`google.api.method_signature`)

**Examples:**

**marketdata.proto - GetAsset:**
```protobuf
// Current
message GetAssetRequest {
  string symbol = 1;
}

// AIP-131 Compliant
message GetAssetRequest {
  // Resource name in format: assets/{asset}
  string name = 1 [(google.api.field_behavior) = REQUIRED];
}

// RPC should include:
rpc GetAsset(GetAssetRequest) returns (Asset) {
  option (google.api.http) = {
    get: "/v1/{name=assets/*}"
  };
  option (google.api.method_signature) = "name";
}
```

**user.proto - GetUser:**
```protobuf
// Current
message GetUserRequest {
  string id = 1;
}

// AIP-131 Compliant
message GetUserRequest {
  // Resource name in format: users/{user}
  string name = 1 [(google.api.field_behavior) = REQUIRED];
}
```

---

#### AIP-132: Standard Methods - List
- **Status**: ⚠️ **Partially Compliant**
- **Issues**:
  1. Missing `parent` field for hierarchical resources
  2. Pagination is implemented correctly (✅ `page_size` and `page_token`)
  3. Missing HTTP annotations
  4. Missing method signature annotations

**Examples:**

**marketdata.proto - ListAssets:**
```protobuf
// Current (Partially compliant)
message ListAssetsRequest {
  int32 page_size = 1;
  string page_token = 2;
}

// AIP-132 Compliant (if assets are top-level)
message ListAssetsRequest {
  // Optional: string parent = 1; // Only if assets have a parent resource
  
  int32 page_size = 2;
  string page_token = 3;
}

// RPC should include:
rpc ListAssets(ListAssetsRequest) returns (ListAssetsResponse) {
  option (google.api.http) = {
    get: "/v1/assets"
  };
  option (google.api.method_signature) = "";
}
```

**transaction.proto - ListTransactions:**
```protobuf
// Current
message ListTransactionsRequest {
  string user_id = 1;  // Should be 'parent'
  int32 page_size = 2;
  string page_token = 3;
  TransactionFilter filter = 4;
}

// AIP-132 Compliant
message ListTransactionsRequest {
  // Parent resource name in format: users/{user}
  string parent = 1 [
    (google.api.field_behavior) = REQUIRED,
    (google.api.resource_reference) = {
      child_type: "portfolioinsights.googleapis.com/Transaction"
    }
  ];
  
  int32 page_size = 2;
  string page_token = 3;
  TransactionFilter filter = 4;
}

// RPC should include:
rpc ListTransactions(ListTransactionsRequest) returns (ListTransactionsResponse) {
  option (google.api.http) = {
    get: "/v1/{parent=users/*}/transactions"
  };
  option (google.api.method_signature) = "parent";
}
```

---

#### AIP-133: Standard Methods - Create
- **Status**: ⚠️ **Partially Compliant**
- **Issues**:
  1. Missing `parent` field in request messages
  2. Resource field should be named after the resource type (e.g., `transaction`, `user`)
  3. Missing HTTP annotations
  4. Missing method signature annotations

**Examples:**

**transaction.proto - CreateTransaction:**
```protobuf
// Current (Non-compliant structure)
message CreateTransactionRequest {
  string user_id = 1;
  string type = 3;
  optional string symbol = 2;
  // ... other fields
}

// AIP-133 Compliant
message CreateTransactionRequest {
  // Parent resource name in format: users/{user}
  string parent = 1 [(google.api.field_behavior) = REQUIRED];
  
  // The transaction to create
  Transaction transaction = 2 [(google.api.field_behavior) = REQUIRED];
  
  // Optional: Client-specified ID
  // string transaction_id = 3;
}

// RPC should include:
rpc CreateTransaction(CreateTransactionRequest) returns (Transaction) {
  option (google.api.http) = {
    post: "/v1/{parent=users/*}/transactions"
    body: "transaction"
  };
  option (google.api.method_signature) = "parent,transaction";
}
```

**user.proto - CreateUser:**
```protobuf
// Current (Non-compliant structure)
message CreateUserRequest {
  string email = 1;
  string username = 2;
  string password = 3;
}

// AIP-133 Compliant
message CreateUserRequest {
  // The user to create
  User user = 1 [(google.api.field_behavior) = REQUIRED];
  
  // Optional: Client-specified ID
  // string user_id = 2;
}

// Note: Password should be handled separately or as INPUT_ONLY field
```

---

#### AIP-134: Standard Methods - Update
- **Status**: ❌ **Major Non-Compliance**
- **Issues**:
  1. Update request should contain the resource object, not individual fields
  2. Missing `google.protobuf.FieldMask update_mask` field
  3. Missing HTTP annotations (should use PATCH)
  4. Missing method signature annotations

**Examples:**

**transaction.proto - UpdateTransaction:**
```protobuf
// Current (Non-compliant)
message UpdateTransactionRequest {
  string id = 1;
  string type = 3;
  optional string symbol = 2;
  optional double quantity = 4;
  // ... many individual fields
}

// AIP-134 Compliant
message UpdateTransactionRequest {
  // The transaction to update.
  // The transaction's `name` field is used to identify the transaction.
  // Format: users/{user}/transactions/{transaction}
  Transaction transaction = 1 [(google.api.field_behavior) = REQUIRED];
  
  // The list of fields to update.
  google.protobuf.FieldMask update_mask = 2;
}

// RPC should include:
rpc UpdateTransaction(UpdateTransactionRequest) returns (Transaction) {
  option (google.api.http) = {
    patch: "/v1/{transaction.name=users/*/transactions/*}"
    body: "transaction"
  };
  option (google.api.method_signature) = "transaction,update_mask";
}
```

**Required import:**
```protobuf
import "google/protobuf/field_mask.proto";
```

---

#### AIP-135: Standard Methods - Delete
- **Status**: ⚠️ **Partially Compliant**
- **Issues**:
  1. Request field should be `name` instead of `id`
  2. Response should be `google.protobuf.Empty` instead of custom response
  3. Missing HTTP annotations
  4. Missing method signature annotations

**Examples:**

**transaction.proto - DeleteTransaction:**
```protobuf
// Current
message DeleteTransactionRequest {
  string id = 1;
}

message DeleteTransactionResponse {
  bool success = 1;
}

// AIP-135 Compliant
message DeleteTransactionRequest {
  // Resource name in format: users/{user}/transactions/{transaction}
  string name = 1 [(google.api.field_behavior) = REQUIRED];
}

// RPC should include:
rpc DeleteTransaction(DeleteTransactionRequest) returns (google.protobuf.Empty) {
  option (google.api.http) = {
    delete: "/v1/{name=users/*/transactions/*}"
  };
  option (google.api.method_signature) = "name";
}
```

**Required import:**
```protobuf
import "google/protobuf/empty.proto";
```

---

#### AIP-203: Field Behavior Documentation
- **Status**: ❌ **Non-Compliant**
- **Issue**: Missing field behavior annotations
- **Expected**: Use `google.api.field_behavior` annotations to indicate REQUIRED, OPTIONAL, OUTPUT_ONLY, INPUT_ONLY, IMMUTABLE

**Examples:**

```protobuf
import "google/api/field_behavior.proto";

message Transaction {
  // Resource name
  string name = 1 [(google.api.field_behavior) = IDENTIFIER];
  
  string user_id = 2 [(google.api.field_behavior) = OUTPUT_ONLY];
  
  string type = 4 [(google.api.field_behavior) = REQUIRED];
  
  optional string symbol = 3;  // OPTIONAL is implicit for optional fields
  
  google.protobuf.Timestamp created_at = 8 [
    (google.api.field_behavior) = OUTPUT_ONLY,
    (google.api.field_behavior) = IMMUTABLE
  ];
}
```

---

#### AIP-136: Custom Methods
- **Status**: ⚠️ **Needs Review**
- **Custom Methods Identified**:
  1. `GetLatestPrice` / `GetLatestPrices` (marketdata.proto)
  2. `GetHistoricalPrices` (marketdata.proto)
  3. `GetLatestCurrencyRate` / `GetHistoricalCurrencyRates` (marketdata.proto)
  4. `GetPortfolioSummary` (portfolio.proto)
  5. `GetPortfolioPerformance` (portfolio.proto)
  6. `BackfillHistory` (portfolio.proto)
  7. `VerifyUser` (user.proto)
  8. `GetOldestTransactionForUser` (transaction.proto)

**Guidance**: Custom methods should:
- Use HTTP POST or GET appropriately
- Have clear naming (verb-first for actions)
- Include proper HTTP annotations
- Consider if they could be standard methods instead

**Example - BackfillHistory:**
```protobuf
// Current (Custom method)
rpc BackfillHistory(BackfillHistoryRequest) returns (BackfillHistoryResponse);

// Should include HTTP annotation:
rpc BackfillHistory(BackfillHistoryRequest) returns (BackfillHistoryResponse) {
  option (google.api.http) = {
    post: "/v1/{parent=users/*}/portfolios:backfillHistory"
    body: "*"
  };
}
```

---

## Detailed File-by-File Analysis

### 1. marketdata.proto

#### Strengths:
- ✅ Excellent documentation
- ✅ Good pagination implementation
- ✅ Proper use of timestamps
- ✅ Batch methods (`GetLatestPrices`)

#### Issues:
| Issue | Severity | AIP | Description |
|-------|----------|-----|-------------|
| No resource names | High | 122 | Uses `id` and `symbol` instead of hierarchical `name` |
| Missing HTTP annotations | Medium | 127 | No `google.api.http` annotations |
| Missing field behavior | Medium | 203 | No field behavior annotations |
| Get method uses symbol | Medium | 131 | `GetAssetRequest` uses `symbol` instead of `name` |

#### Recommendations:
1. Introduce resource names: `assets/{asset}`
2. Add HTTP/gRPC transcoding annotations
3. Add field behavior annotations
4. Consider if `GetAssetBySymbol` should be a custom method while `GetAsset` uses resource name

---

### 2. portfolio.proto

#### Strengths:
- ✅ Good documentation
- ✅ Clear domain modeling
- ✅ Proper timestamp usage

#### Issues:
| Issue | Severity | AIP | Description |
|-------|----------|-----|-------------|
| No resource names | High | 122 | Uses `user_id` instead of hierarchical resource names |
| No parent field in requests | High | 132 | List/Get methods should use `parent` or `name` |
| Missing HTTP annotations | Medium | 127 | No `google.api.http` annotations |
| Missing field behavior | Medium | 203 | No field behavior annotations |
| Custom method naming | Low | 136 | `BackfillHistory` could use better HTTP annotation |

#### Recommendations:
1. Introduce resource hierarchy: `users/{user}/portfolios/{portfolio}`
2. Add `parent` field to request messages
3. Add HTTP annotations for all RPCs
4. Add field behavior annotations

---

### 3. transaction.proto

#### Strengths:
- ✅ Comprehensive documentation
- ✅ Good filter pattern
- ✅ Proper pagination
- ✅ Support for multiple transaction types

#### Issues:
| Issue | Severity | AIP | Description |
|-------|----------|-----|-------------|
| No resource names | High | 122 | Uses `id` instead of hierarchical `name` |
| Update uses individual fields | High | 134 | Should use resource + field mask pattern |
| Create uses individual fields | High | 133 | Should use resource object pattern |
| No parent field | High | 132 | `ListTransactions` uses `user_id` instead of `parent` |
| Delete returns custom response | Medium | 135 | Should return `google.protobuf.Empty` |
| Missing HTTP annotations | Medium | 127 | No `google.api.http` annotations |
| Missing field behavior | Medium | 203 | No field behavior annotations |

#### Recommendations:
1. **Critical**: Refactor to use resource names: `users/{user}/transactions/{transaction}`
2. **Critical**: Refactor `UpdateTransactionRequest` to use resource + field mask
3. **Critical**: Refactor `CreateTransactionRequest` to use Transaction object
4. Change `DeleteTransactionResponse` to `google.protobuf.Empty`
5. Add HTTP annotations
6. Add field behavior annotations

---

### 4. user.proto

#### Strengths:
- ✅ Simple and clear
- ✅ Good documentation

#### Issues:
| Issue | Severity | AIP | Description |
|-------|----------|-----|-------------|
| No resource names | High | 122 | Uses `id` instead of `name` |
| Create uses individual fields | High | 133 | Should use User object pattern |
| Password in response | High | Security | `GetUserResponse` could expose password if not careful |
| Missing HTTP annotations | Medium | 127 | No `google.api.http` annotations |
| Missing field behavior | Medium | 203 | No field behavior annotations |
| VerifyUser method | Low | 136 | Custom method needs HTTP annotation |

#### Recommendations:
1. Introduce resource names: `users/{user}`
2. Refactor `CreateUserRequest` to use User object
3. Mark password as `INPUT_ONLY` and ensure it's never returned
4. Add HTTP annotations
5. Add field behavior annotations
6. Consider authentication best practices for `VerifyUser`

---

## Priority Recommendations

### 🔴 High Priority (Breaking Changes)

1. **Implement Resource Names (AIP-122)**
   - Add `name` field to all resources
   - Use hierarchical naming: `users/{user}/transactions/{transaction}`
   - Update all request messages to use `name` or `parent`

2. **Refactor Update Methods (AIP-134)**
   - Use resource object + field mask pattern
   - Import `google/protobuf/field_mask.proto`

3. **Refactor Create Methods (AIP-133)**
   - Use resource object pattern
   - Add `parent` field where appropriate

4. **Fix Delete Methods (AIP-135)**
   - Return `google.protobuf.Empty`
   - Use `name` field in request

### 🟡 Medium Priority (Enhancements)

5. **Add HTTP/gRPC Transcoding (AIP-127)**
   - Add `google.api.http` annotations to all RPCs
   - Add `google.api.method_signature` annotations
   - Import `google/api/annotations.proto`

6. **Add Field Behavior Annotations (AIP-203)**
   - Mark required fields with `REQUIRED`
   - Mark output-only fields with `OUTPUT_ONLY`
   - Mark immutable fields with `IMMUTABLE`
   - Mark input-only fields (like passwords) with `INPUT_ONLY`
   - Import `google/api/field_behavior.proto`

7. **Add Resource Type Annotations (AIP-123)**
   - Define resource types with `google.api.resource`
   - Add resource references with `google.api.resource_reference`
   - Import `google/api/resource.proto`

### 🟢 Low Priority (Polish)

8. **Review Custom Methods (AIP-136)**
   - Ensure custom methods have proper HTTP annotations
   - Consider if any should be standard methods

9. **Add Filtering Guidance (AIP-160)**
   - Document filter syntax in comments
   - Consider using standard filter expressions

10. **Consider Long-Running Operations (AIP-151)**
    - `BackfillHistory` might benefit from LRO pattern
    - Import `google/longrunning/operations.proto` if needed

---

## Required Imports for Full Compliance

```protobuf
import "google/api/annotations.proto";      // For HTTP annotations
import "google/api/client.proto";           // For method signatures
import "google/api/field_behavior.proto";   // For field behavior
import "google/api/resource.proto";         // For resource definitions
import "google/protobuf/empty.proto";       // For Delete responses
import "google/protobuf/field_mask.proto";  // For Update operations
import "google/protobuf/timestamp.proto";   // Already used ✅
```

---

## Migration Strategy

Given the breaking nature of many required changes, consider:

1. **Versioning**: Create a new API version (e.g., `v2`) with AIP-compliant definitions
2. **Dual Support**: Maintain both old and new APIs during transition
3. **Gradual Migration**: Start with new services, migrate existing ones over time
4. **Client Library Impact**: Coordinate with client library users on migration timeline

---

## Conclusion

The current proto files are well-documented and functional but deviate significantly from Google's AIP standards. The most critical gaps are:

1. **Lack of hierarchical resource names** (AIP-122)
2. **Non-standard Update method pattern** (AIP-134)
3. **Non-standard Create method pattern** (AIP-133)
4. **Missing HTTP annotations** (AIP-127)
5. **Missing field behavior annotations** (AIP-203)

Addressing these issues would significantly improve:
- API consistency and predictability
- Client library generation quality
- Interoperability with Google Cloud tools
- Long-term maintainability
- Developer experience

The recommended approach is to create a new API version with full AIP compliance while maintaining backward compatibility with the current version during a transition period.
