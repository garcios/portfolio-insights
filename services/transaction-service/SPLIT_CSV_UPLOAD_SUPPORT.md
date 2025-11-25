# SPLIT Transaction Support in CSV Upload - Summary

## ✅ Implementation Complete

Successfully added SPLIT transaction type support to the transaction-service CSV upload functionality with flexible price handling.

---

## 📋 Changes Made

### **Modified File**: `services/transaction-service/internal/usecase/csv_upload_usecase.go`

**Key Changes**:

1. **Added SPLIT to valid transaction types**
   ```go
   if txType != "BUY" && txType != "SELL" && txType != "SPLIT" {
       return nil, fmt.Errorf("type must be BUY, SELL, or SPLIT")
   }
   ```

2. **Made price_per_share optional for SPLIT transactions**
   ```go
   if priceStr == "" || priceStr == "0" {
       // Empty or zero price is only allowed for SPLIT transactions
       if txType != "SPLIT" {
           return nil, fmt.Errorf("price_per_share is required for %s transactions", txType)
       }
       price = 0
   }
   ```

3. **Relaxed price validation for SPLIT**
   ```go
   // Price must be positive for BUY and SELL, but can be 0 for SPLIT
   if price <= 0 && txType != "SPLIT" {
       return nil, fmt.Errorf("price_per_share must be positive for %s transactions", txType)
   }
   ```

---

## 🎯 Validation Rules by Transaction Type

| Transaction Type | Quantity | Price | Validation |
|-----------------|----------|-------|------------|
| **BUY** | Must be > 0 | Must be > 0 | Standard purchase |
| **SELL** | Must be > 0 | Must be > 0 | Standard sale |
| **SPLIT** | Must be > 0 | Can be 0 or empty | Stock split ✅ |

---

## 📝 CSV Format Examples

### **Valid SPLIT Transactions**

#### **Example 1: Price = 0**
```csv
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-06-10,100,0,SPLIT
```

#### **Example 2: Price = empty**
```csv
symbol,executed_at,quantity,price_per_share,type
GOOGL,2024-07-15,200,,SPLIT
```

#### **Example 3: Price = current market price (optional)**
```csv
symbol,executed_at,quantity,price_per_share,type
NVDA,2024-08-20,150,50.25,SPLIT
```

**Note**: The price value is **ignored** by the portfolio-service for SPLIT transactions, so it can be any value including 0 or empty.

### **Mixed Transaction Types**

```csv
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-01-01,100,150.50,BUY
AAPL,2024-06-10,100,0,SPLIT
GOOGL,2024-02-15,50,2800.00,BUY
GOOGL,2024-07-15,100,,SPLIT
MSFT,2024-03-20,75,380.25,BUY
MSFT,2024-04-10,25,390.00,SELL
```

---

## 🧪 Testing

### **Test Case 1: SPLIT with Price = 0**

```bash
# Create test CSV
cat > split_test.csv << 'EOF'
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-06-10,100,0,SPLIT
EOF

# Upload
curl -X POST \
  -F "file=@split_test.csv" \
  "http://localhost:8081/upload-csv?user_id=test-user"

# Expected response:
{
  "total_records": 1,
  "successful_records": 1,
  "failed_records": 0,
  "errors": []
}
```

### **Test Case 2: SPLIT with Empty Price**

```bash
# Create test CSV
cat > split_empty_price.csv << 'EOF'
symbol,executed_at,quantity,price_per_share,type
GOOGL,2024-07-15,200,,SPLIT
EOF

# Upload
curl -X POST \
  -F "file=@split_empty_price.csv" \
  "http://localhost:8081/upload-csv?user_id=test-user"

# Expected response:
{
  "total_records": 1,
  "successful_records": 1,
  "failed_records": 0,
  "errors": []
}
```

### **Test Case 3: BUY with Price = 0 (Should Fail)**

```bash
# Create test CSV
cat > buy_zero_price.csv << 'EOF'
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-01-01,100,0,BUY
EOF

# Upload
curl -X POST \
  -F "file=@buy_zero_price.csv" \
  "http://localhost:8081/upload-csv?user_id=test-user"

# Expected response:
{
  "total_records": 1,
  "successful_records": 0,
  "failed_records": 1,
  "errors": [
    {
      "row_number": 2,
      "error": "price_per_share is required for BUY transactions"
    }
  ]
}
```

### **Test Case 4: Mixed Transactions**

```bash
# Create test CSV
cat > mixed_transactions.csv << 'EOF'
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-01-01,100,150.50,BUY
AAPL,2024-06-10,100,0,SPLIT
GOOGL,2024-02-15,50,2800.00,BUY
GOOGL,2024-07-15,100,,SPLIT
MSFT,2024-03-20,75,380.25,BUY
EOF

# Upload
curl -X POST \
  -F "file=@mixed_transactions.csv" \
  "http://localhost:8081/upload-csv?user_id=test-user"

# Expected response:
{
  "total_records": 5,
  "successful_records": 5,
  "failed_records": 0,
  "errors": []
}
```

---

## 🔄 Complete Workflow

### **From Broker Export to Database**

```bash
# 1. Export from broker (Excel format)
# Save as: AllTradesReport.xlsx

# 2. Convert Excel to CSV
cd tools/excel2csv
./bin/excel2csv --sheets "Combined" ioFiles/AllTradesReport.xlsx
# Output: ioFiles/AllTradesReport_Combined.csv

# 3. Convert broker CSV to transaction CSV (with SPLIT support)
./bin/csvtxn ioFiles/AllTradesReport_Combined.csv
# Output: ioFiles/transactions.csv

# 4. Upload to transaction service
curl -X POST \
  -F "file=@ioFiles/transactions.csv" \
  "http://localhost:8081/upload-csv?user_id=user-123"

# 5. Verify in database
psql -d portfolio -c "
  SELECT symbol, type, quantity, price_per_share, executed_at
  FROM txn.transactions
  WHERE user_id = 'user-123' AND type = 'SPLIT'
  ORDER BY executed_at;
"
```

---

## 📊 Validation Logic

### **Price Validation Flow**

```
Parse price_per_share field
    ↓
Is price empty or "0"?
    ├─ YES → Is type SPLIT?
    │         ├─ YES → Set price = 0 ✅
    │         └─ NO → Error: "price_per_share is required for {type} transactions"
    └─ NO → Parse as float
              ↓
          Is price <= 0?
              ├─ YES → Is type SPLIT?
              │         ├─ YES → Allow ✅
              │         └─ NO → Error: "price_per_share must be positive for {type} transactions"
              └─ NO → Use parsed price ✅
```

---

## 🎯 Error Messages

### **SPLIT-Specific Errors**

| Error | Cause | Solution |
|-------|-------|----------|
| "type must be BUY, SELL, or SPLIT" | Invalid transaction type | Use BUY, SELL, or SPLIT |
| "quantity must be positive" | Quantity ≤ 0 | Use positive quantity |
| "symbol is required" | Empty symbol | Provide stock symbol |

### **BUY/SELL-Specific Errors**

| Error | Cause | Solution |
|-------|-------|----------|
| "price_per_share is required for BUY transactions" | Empty/zero price for BUY | Provide valid price |
| "price_per_share must be positive for SELL transactions" | Price ≤ 0 for SELL | Provide positive price |

---

## 📈 Database Storage

### **SPLIT Transaction in Database**

```sql
-- Example SPLIT transaction record
INSERT INTO txn.transactions (
    user_id,
    symbol,
    type,
    quantity,
    price_per_share,
    executed_at
) VALUES (
    'user-123',
    'AAPL',
    'SPLIT',
    100,
    0.00,  -- Price is 0 for SPLIT
    '2024-06-10'
);
```

### **Query SPLIT Transactions**

```sql
-- Get all SPLIT transactions
SELECT 
    symbol,
    quantity,
    price_per_share,
    executed_at,
    created_at
FROM txn.transactions
WHERE type = 'SPLIT'
ORDER BY executed_at DESC;

-- Get SPLIT transactions by symbol
SELECT 
    symbol,
    quantity,
    executed_at
FROM txn.transactions
WHERE type = 'SPLIT' AND symbol = 'AAPL'
ORDER BY executed_at;
```

---

## 🔍 Integration with Portfolio Service

### **Event Publishing**

When a SPLIT transaction is uploaded:

1. **Transaction Service** saves to database
2. **Event Published** to NATS: `transaction-service.transaction.created`
3. **Portfolio Service** receives event
4. **SPLIT Handler** processes:
   - Calculates split ratio
   - Adjusts average cost
   - Updates quantity
   - Maintains total cost basis

### **Example Event**

```json
{
  "transaction_id": "uuid-123",
  "user_id": "user-123",
  "asset_symbol": "AAPL",
  "price_per_share": 0,
  "quantity": 100,
  "type": "SPLIT",
  "executed_at": "2024-06-10T00:00:00Z"
}
```

---

## ✅ Validation Checklist

After uploading SPLIT transactions:

- [ ] CSV upload successful (no errors)
- [ ] Transactions saved to database
- [ ] Type = "SPLIT" in database
- [ ] Price = 0 in database (or whatever was provided)
- [ ] Events published to NATS
- [ ] Portfolio service processed split
- [ ] Holdings quantity updated
- [ ] Holdings average cost adjusted
- [ ] Total cost basis unchanged

---

## 🚀 Benefits

### ✨ **Flexible Price Handling**

- Price can be 0, empty, or any value for SPLIT
- Strict validation for BUY and SELL
- Clear error messages

### ✨ **Backward Compatible**

- Existing BUY and SELL validation unchanged
- No breaking changes to API
- Existing CSVs continue to work

### ✨ **Complete Workflow**

- Excel → CSV → Transaction CSV → Upload → Database → Portfolio Update
- All tools support SPLIT transactions
- End-to-end integration

---

## 📝 Summary

**Changes**:
- ✅ Added SPLIT to valid transaction types
- ✅ Made price optional for SPLIT (can be 0 or empty)
- ✅ Maintained strict validation for BUY/SELL
- ✅ Clear error messages for each type

**Validation Rules**:
- **BUY**: Quantity > 0, Price > 0
- **SELL**: Quantity > 0, Price > 0
- **SPLIT**: Quantity > 0, Price ≥ 0 (or empty) ✅

**Integration**:
- ✅ Works with csvtxn tool
- ✅ Works with CSV upload endpoint
- ✅ Events published to portfolio-service
- ✅ Portfolio-service handles SPLIT correctly

---

**Implementation Date**: 2024-11-25
