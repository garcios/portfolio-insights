.PHONY: run-gateway run-user run-portfolio

run-gateway:
	cd apps/gateway && go run cmd/server/main.go

run-user:
	cd services/user-service && go run cmd/server/main.go

run-portfolio:
	cd services/portfolio-service && go run cmd/server/main.go

docker-up:
	docker-compose -f deployments/docker-compose/docker-compose.yml up --build

proto-gen:
	@mkdir -p services/user-service/proto/user
	@mkdir -p services/transaction-service/proto/transaction
	protoc --go_out=services/user-service --go_opt=paths=source_relative \
    --go-grpc_out=services/user-service --go-grpc_opt=paths=source_relative \
    proto/user/user.proto
	protoc --go_out=services/transaction-service --go_opt=paths=source_relative \
    --go-grpc_out=services/transaction-service --go-grpc_opt=paths=source_relative \
    proto/transaction/transaction.proto
	@mkdir -p services/marketdata-service/proto/marketdata
	protoc --go_out=services/marketdata-service --go_opt=paths=source_relative \
    --go-grpc_out=services/marketdata-service --go-grpc_opt=paths=source_relative \
    proto/marketdata/marketdata.proto
	@mkdir -p services/portfolio-service/proto/portfolio
	protoc --go_out=services/portfolio-service --go_opt=paths=source_relative \
    --go-grpc_out=services/portfolio-service --go-grpc_opt=paths=source_relative \
    proto/portfolio/portfolio.proto
