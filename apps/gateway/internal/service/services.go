package service

import (
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
)

// Services is a container struct that holds all application services.
type Services struct {
	Transaction *TransactionService
}

// NewServices is the constructor for the Services container.
// It takes raw gRPC clients and uses them to create the service layer instances.
func NewServices(
	transactionClient transactionpb.TransactionServiceClient,
) *Services {
	return &Services{
		Transaction: NewTransactionService(transactionClient),
	}
}
