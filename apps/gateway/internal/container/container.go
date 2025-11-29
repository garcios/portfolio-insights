package container

import (
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/grpc"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/usecase"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
)

// Container holds all application dependencies
type Container struct {
	// Gateways
	UserGateway        gateway.UserGateway
	PortfolioGateway   gateway.PortfolioGateway
	TransactionGateway gateway.TransactionGateway

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
) *Container {
	// Initialize gateways
	userGateway := grpc.NewUserGRPCGateway(userClient)
	portfolioGateway := grpc.NewPortfolioGRPCGateway(portfolioClient)
	transactionGateway := grpc.NewTransactionGRPCGateway(transactionClient)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userGateway)
	portfolioUseCase := usecase.NewPortfolioUseCase(portfolioGateway)
	transactionUseCase := usecase.NewTransactionUseCase(transactionGateway)

	return &Container{
		UserGateway:        userGateway,
		PortfolioGateway:   portfolioGateway,
		TransactionGateway: transactionGateway,
		UserUseCase:        userUseCase,
		PortfolioUseCase:   portfolioUseCase,
		TransactionUseCase: transactionUseCase,
	}
}
