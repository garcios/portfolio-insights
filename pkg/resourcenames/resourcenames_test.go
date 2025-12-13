package resourcenames

import (
	"testing"
)

func TestUserName(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		want   string
	}{
		{"basic", "123", "users/123"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", "users/550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserName(tt.userID); got != tt.want {
				t.Errorf("UserName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUserName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid", "users/123", "123", false},
		{"valid uuid", "users/550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid format", "user/123", "", true},
		{"missing id", "users/", "", true},
		{"too many parts", "users/123/extra", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUserName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUserName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseUserName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionName(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		transactionID string
		want          string
	}{
		{"basic", "123", "txn-456", "users/123/transactions/txn-456"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", "txn-789", "users/550e8400-e29b-41d4-a716-446655440000/transactions/txn-789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TransactionName(tt.userID, tt.transactionID); got != tt.want {
				t.Errorf("TransactionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTransactionName(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantUserID        string
		wantTransactionID string
		wantErr           bool
	}{
		{"valid", "users/123/transactions/txn-456", "123", "txn-456", false},
		{"invalid format", "users/123/transaction/txn-456", "", "", true},
		{"missing parts", "users/123", "", "", true},
		{"too many parts", "users/123/transactions/txn-456/extra", "", "", true},
		{"empty", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, gotTransactionID, err := ParseTransactionName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTransactionName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ParseTransactionName() userID = %v, want %v", gotUserID, tt.wantUserID)
			}
			if gotTransactionID != tt.wantTransactionID {
				t.Errorf("ParseTransactionName() transactionID = %v, want %v", gotTransactionID, tt.wantTransactionID)
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		name    string
		assetID string
		want    string
	}{
		{"basic", "aapl", "assets/aapl"},
		{"numeric", "123", "assets/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AssetName(tt.assetID); got != tt.want {
				t.Errorf("AssetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAssetName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid", "assets/aapl", "aapl", false},
		{"invalid format", "asset/aapl", "", true},
		{"missing id", "assets/", "", true},
		{"too many parts", "assets/aapl/extra", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAssetName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAssetName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAssetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortfolioName(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		want   string
	}{
		{"basic", "123", "users/123/portfolio"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", "users/550e8400-e29b-41d4-a716-446655440000/portfolio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PortfolioName(tt.userID); got != tt.want {
				t.Errorf("PortfolioName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePortfolioName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid", "users/123/portfolio", "123", false},
		{"invalid format", "users/123/portfolios", "", true},
		{"missing parts", "users/123", "", true},
		{"too many parts", "users/123/portfolio/extra", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePortfolioName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortfolioName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePortfolioName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHoldingName(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		holdingID string
		want      string
	}{
		{"basic", "123", "holding-456", "users/123/holdings/holding-456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HoldingName(tt.userID, tt.holdingID); got != tt.want {
				t.Errorf("HoldingName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHoldingName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantUserID    string
		wantHoldingID string
		wantErr       bool
	}{
		{"valid", "users/123/holdings/holding-456", "123", "holding-456", false},
		{"invalid format", "users/123/holding/holding-456", "", "", true},
		{"missing parts", "users/123", "", "", true},
		{"empty", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, gotHoldingID, err := ParseHoldingName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHoldingName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ParseHoldingName() userID = %v, want %v", gotUserID, tt.wantUserID)
			}
			if gotHoldingID != tt.wantHoldingID {
				t.Errorf("ParseHoldingName() holdingID = %v, want %v", gotHoldingID, tt.wantHoldingID)
			}
		})
	}
}
