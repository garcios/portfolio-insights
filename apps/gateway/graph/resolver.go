package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
)

type Resolver struct {
	PortfolioClient   portfoliopb.PortfolioServiceClient
	UserClient        userpb.UserServiceClient
	TransactionClient transactionpb.TransactionServiceClient
}
