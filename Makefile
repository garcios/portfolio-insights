.PHONY: run-gateway run-user run-portfolio tidy-all

run-gateway:
	cd apps/gateway && HYDRA_PUBLIC_URL=http://localhost:4444 go run cmd/server/main.go

run-user:
	cd services/user-service && go run cmd/server/main.go

run-portfolio:
	cd services/portfolio-service && go run cmd/server/main.go

podman-up:
	podman-compose -f deployments/docker-compose/docker-compose.yml up --build

podman-down:
	podman-compose -f deployments/docker-compose/docker-compose.yml down

podman-logs:
	podman-compose -f deployments/docker-compose/docker-compose.yml logs -f

podman-build-gateway:
	podman-compose -f deployments/docker-compose/docker-compose.yml up -d --build gateway

monitoring-up:
	./deployments/monitoring/start-monitoring.sh

monitoring-down:
	./deployments/monitoring/stop-monitoring.sh

monitoring-logs:
	podman logs -f prometheus

hydra-build:
	podman-compose -f deployments/docker-compose/docker-compose.hydra.yml build --no-cache login-consent-provider

hydra-up:
	podman-compose -f deployments/docker-compose/docker-compose.hydra.yml up -d

hydra-down:
	podman-compose -f deployments/docker-compose/docker-compose.hydra.yml down 

hydra-logs:
	podman-compose -f deployments/docker-compose/docker-compose.hydra.yml logs -f

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

# Testing targets
test-portfolio:
	cd services/portfolio-service && go test ./internal/... -v

test-portfolio-coverage:
	cd services/portfolio-service && go test ./internal/... -cover -coverprofile=coverage.out
	cd services/portfolio-service && go tool cover -html=coverage.out

test-portfolio-race:
	cd services/portfolio-service && go test ./internal/... -race

test-user:
	cd services/user-service && go test ./internal/... -v

test-transaction:
	cd services/transaction-service && go test ./internal/... -v

test-marketdata:
	cd services/marketdata-service && go test ./internal/... -v

test-login-consent-provider:
	cd apps/login-consent-provider && go test ./internal/... -v

test-gateway:
	cd apps/gateway && go test ./internal/... -v

test-all:
	cd services/user-service && go test ./internal/... -v
	cd services/transaction-service && go test ./internal/... -v
	cd services/portfolio-service && go test ./internal/... -v
	cd services/marketdata-service && go test ./internal/... -v
	cd apps/login-consent-provider && go test ./... -v
	cd apps/gateway && go test ./internal/... -v

podman-prune:
	podman system prune -a --volumes

tidy-all:	
	cd services/user-service && go mod tidy
	cd services/transaction-service && go mod tidy
	cd services/portfolio-service && go mod tidy
	cd services/marketdata-service && go mod tidy		
	cd apps/login-consent-provider && go mod tidy
	cd apps/gateway && go mod tidy