// Package container provides dependency injection.
package container

import (
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/grpc"
	httpgw "github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/http"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/usecase"
	marketdatapb "github.com/garcios/portfolio-insights/services/marketdata-service/marketdata"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
)

// Container holds all application dependencies
type Container struct {
	// Gateways
	UserGateway            gateway.UserGateway
	PortfolioGateway       gateway.PortfolioGateway
	TransactionGateway     gateway.TransactionGateway
	TransactionFileGateway gateway.TransactionFileGateway
	MarketDataGateway      gateway.MarketDataGateway

	// Use Cases
	UserUseCase        *usecase.UserUseCase
	PortfolioUseCase   *usecase.PortfolioUseCase
	TransactionUseCase *usecase.TransactionUseCase
}

// NewContainer creates a new dependency injection container
func NewContainer(
	userClient userpb.UserServiceClient,
	portfolioClient portfoliopb.PortfolioServiceClient,
	transactionClient transactionpb.TransactionServiceClient,
	transactionServiceURL string,
	marketDataClient marketdatapb.MarketDataServiceClient,
) *Container {
	// Initialize gateways
	userGateway := grpc.NewUserGRPCGateway(userClient)
	portfolioGateway := grpc.NewPortfolioGRPCGateway(portfolioClient)
	transactionGateway := grpc.NewTransactionGRPCGateway(transactionClient)
	transactionFileGateway := httpgw.NewTransactionHTTPGateway(transactionServiceURL)
	marketDataGateway := grpc.NewMarketDataGRPCGateway(marketDataClient)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userGateway)
	portfolioUseCase := usecase.NewPortfolioUseCase(portfolioGateway)
	transactionUseCase := usecase.NewTransactionUseCase(transactionGateway, transactionFileGateway)

	return &Container{
		UserGateway:            userGateway,
		PortfolioGateway:       portfolioGateway,
		TransactionGateway:     transactionGateway,
		TransactionFileGateway: transactionFileGateway,
		MarketDataGateway:      marketDataGateway,
		UserUseCase:            userUseCase,
		PortfolioUseCase:       portfolioUseCase,
		TransactionUseCase:     transactionUseCase,
	}
}
