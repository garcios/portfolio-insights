.PHONY: run-gateway run-local-user-service run-local-portfolio-service run-local-transaction-service run-local-marketdata-service tidy-all build-postgres run-login-consent-provider

run-gateway:
	cd apps/gateway && HYDRA_PUBLIC_URL=http://localhost:4444 APP_ENV=local APP_CONFIG_PATH=./configs go run cmd/server/main.go

run-login-consent-provider:
	cd apps/login-consent-provider && APP_ENV=local APP_CONFIG_PATH=./configs go run cmd/server/main.go

run-local-user-service: 
	cd services/user-service &&  APP_ENV=local APP_CONFIG_PATH=./configs go run cmd/server/main.go

run-local-portfolio-service: 
	cd services/portfolio-service &&  APP_ENV=local APP_CONFIG_PATH=./configs go run cmd/server/main.go

run-local-transaction-service: 
	cd services/transaction-service &&  APP_ENV=local APP_CONFIG_PATH=./configs go run cmd/server/main.go

run-local-marketdata-service: 
	cd services/marketdata-service &&  APP_ENV=local APP_CONFIG_PATH=./configs go run cmd/server/main.go

services-up:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build

services-down:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml down

services-logs:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml logs -f

build-gateway:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build gateway

build-user-service:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build user-service

build-portfolio-service:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build portfolio-service

build-transaction-service:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build transaction-service

build-marketdata-service:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build marketdata-service

build-postgres:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build postgres

build-redis:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build redis

build-minio:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build minio

build-nats:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build nats

build-migrations:
	podman-compose --env-file deployments/docker-compose-services/.env -f deployments/docker-compose-services/docker-compose.yml up -d --build migrations

build-infra: build-postgres build-redis build-minio build-nats build-migrations

monitoring-up:
	./deployments/monitoring/start-monitoring.sh	

monitoring-down:
	./deployments/monitoring/stop-monitoring.sh

monitoring-logs:
	podman logs -f prometheus

hydra-build:
	podman-compose -f deployments/docker-compose-hydra/docker-compose.hydra.yml build --no-cache login-consent-provider

hydra-up:
	podman-compose -f deployments/docker-compose-hydra/docker-compose.hydra.yml up -d

hydra-down:
	podman-compose -f deployments/docker-compose-hydra/docker-compose.hydra.yml down 

hydra-logs:
	podman-compose -f deployments/docker-compose-hydra/docker-compose.hydra.yml logs -f

proto-gen:
	@mkdir -p services/user-service/proto/user
	@mkdir -p services/transaction-service/proto/transaction
	protoc -I proto --go_out=services/user-service --go_opt=paths=source_relative \
    --go-grpc_out=services/user-service --go-grpc_opt=paths=source_relative \
    proto/user/user.proto
	protoc -I proto --go_out=services/transaction-service --go_opt=paths=source_relative \
    --go-grpc_out=services/transaction-service --go-grpc_opt=paths=source_relative \
    proto/transaction/transaction.proto
	@mkdir -p services/marketdata-service/proto/marketdata
	protoc -I proto --go_out=services/marketdata-service --go_opt=paths=source_relative \
    --go-grpc_out=services/marketdata-service --go-grpc_opt=paths=source_relative \
    proto/marketdata/marketdata.proto
	@mkdir -p services/portfolio-service/proto/portfolio
	protoc -I proto --go_out=services/portfolio-service --go_opt=paths=source_relative \
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
	cd apps/login-consent-provider && go test ./... -v

test-gateway:
	cd apps/gateway && go test ./... -v

test-all:
	cd services/user-service && go test ./internal/... -v
	cd services/transaction-service && go test ./internal/... -v
	cd services/portfolio-service && go test ./internal/... -v
	cd services/marketdata-service && go test ./internal/... -v
	cd apps/login-consent-provider && go test ./... -v
	cd apps/gateway && go test ./... -v

lint:
	cd services/user-service && golangci-lint run ./...
	cd services/transaction-service && golangci-lint run ./...
	cd services/portfolio-service && golangci-lint run ./...
	cd services/marketdata-service && golangci-lint run ./...
	cd apps/login-consent-provider && golangci-lint run ./...
	cd apps/gateway && golangci-lint run ./...

prune:
	podman system prune -a --volumes

tidy-all:	
	cd services/user-service && go mod tidy
	cd services/transaction-service && go mod tidy
	cd services/portfolio-service && go mod tidy
	cd services/marketdata-service && go mod tidy		
	cd apps/login-consent-provider && go mod tidy
	cd apps/gateway && go mod tidy