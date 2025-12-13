// Package resourcenames provides utilities for constructing and parsing
// AIP-compliant resource names across all services.
package resourcenames

import (
	"fmt"
	"strings"
)

// User resource name helpers

// UserName constructs a user resource name from a user ID.
// Format: users/{user}
func UserName(userID string) string {
	return fmt.Sprintf("users/%s", userID)
}

// ParseUserName parses a user resource name and returns the user ID.
// Expected format: users/{user}
func ParseUserName(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] != "users" {
		return "", fmt.Errorf("invalid user name: %s (expected format: users/{user})", name)
	}
	if parts[1] == "" {
		return "", fmt.Errorf("invalid user name: %s (user ID cannot be empty)", name)
	}
	return parts[1], nil
}

// Transaction resource name helpers

// TransactionName constructs a transaction resource name from user ID and transaction ID.
// Format: users/{user}/transactions/{transaction}
func TransactionName(userID, transactionID string) string {
	return fmt.Sprintf("users/%s/transactions/%s", userID, transactionID)
}

// ParseTransactionName parses a transaction resource name and returns the user ID and transaction ID.
// Expected format: users/{user}/transactions/{transaction}
func ParseTransactionName(name string) (userID, transactionID string, err error) {
	parts := strings.Split(name, "/")
	if len(parts) != 4 || parts[0] != "users" || parts[2] != "transactions" {
		return "", "", fmt.Errorf("invalid transaction name: %s (expected format: users/{user}/transactions/{transaction})", name)
	}
	return parts[1], parts[3], nil
}

// ParseTransactionParent parses a transaction parent resource name and returns the user ID.
// Expected format: users/{user}
func ParseTransactionParent(parent string) (string, error) {
	return ParseUserName(parent)
}

// Asset resource name helpers

// AssetName constructs an asset resource name from an asset ID.
// Format: assets/{asset}
func AssetName(assetID string) string {
	return fmt.Sprintf("assets/%s", assetID)
}

// ParseAssetName parses an asset resource name and returns the asset ID.
// Expected format: assets/{asset}
func ParseAssetName(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] != "assets" {
		return "", fmt.Errorf("invalid asset name: %s (expected format: assets/{asset})", name)
	}
	if parts[1] == "" {
		return "", fmt.Errorf("invalid asset name: %s (asset ID cannot be empty)", name)
	}
	return parts[1], nil
}

// Portfolio resource name helpers

// PortfolioName constructs a portfolio resource name from a user ID.
// Format: users/{user}/portfolio (singleton)
func PortfolioName(userID string) string {
	return fmt.Sprintf("users/%s/portfolio", userID)
}

// ParsePortfolioName parses a portfolio resource name and returns the user ID.
// Expected format: users/{user}/portfolio
func ParsePortfolioName(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 3 || parts[0] != "users" || parts[2] != "portfolio" {
		return "", fmt.Errorf("invalid portfolio name: %s (expected format: users/{user}/portfolio)", name)
	}
	return parts[1], nil
}

// Holding resource name helpers

// HoldingName constructs a holding resource name from user ID and holding ID.
// Format: users/{user}/holdings/{holding}
func HoldingName(userID, holdingID string) string {
	return fmt.Sprintf("users/%s/holdings/%s", userID, holdingID)
}

// ParseHoldingName parses a holding resource name and returns the user ID and holding ID.
// Expected format: users/{user}/holdings/{holding}
func ParseHoldingName(name string) (userID, holdingID string, err error) {
	parts := strings.Split(name, "/")
	if len(parts) != 4 || parts[0] != "users" || parts[2] != "holdings" {
		return "", "", fmt.Errorf("invalid holding name: %s (expected format: users/{user}/holdings/{holding})", name)
	}
	return parts[1], parts[3], nil
}

// ParseHoldingParent parses a holding parent resource name and returns the user ID.
// Expected format: users/{user}
func ParseHoldingParent(parent string) (string, error) {
	return ParseUserName(parent)
}

// CurrencyRate resource name helpers

// CurrencyRateName constructs a currency rate resource name from a currency rate ID.
// Format: currencyRates/{currency_rate}
func CurrencyRateName(currencyRateID string) string {
	return fmt.Sprintf("currencyRates/%s", currencyRateID)
}

// ParseCurrencyRateName parses a currency rate resource name and returns the currency rate ID.
// Expected format: currencyRates/{currency_rate}
func ParseCurrencyRateName(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] != "currencyRates" {
		return "", fmt.Errorf("invalid currency rate name: %s (expected format: currencyRates/{currency_rate})", name)
	}
	return parts[1], nil
}

// Validation helpers

// ValidateUserID validates that a user ID is not empty.
func ValidateUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	return nil
}

// ValidateTransactionID validates that a transaction ID is not empty.
func ValidateTransactionID(transactionID string) error {
	if transactionID == "" {
		return fmt.Errorf("transaction ID cannot be empty")
	}
	return nil
}

// ValidateAssetID validates that an asset ID is not empty.
func ValidateAssetID(assetID string) error {
	if assetID == "" {
		return fmt.Errorf("asset ID cannot be empty")
	}
	return nil
}
