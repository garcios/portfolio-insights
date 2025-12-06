FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

COPY . .

# Download dependencies for the gateway module
WORKDIR /app/apps/gateway
RUN go mod download

# Install gqlgen and generate GraphQL code
WORKDIR /app/apps/gateway
RUN go install github.com/99designs/gqlgen@v0.17.45
RUN go run github.com/99designs/gqlgen generate

# Build the server
WORKDIR /app
RUN go build -o /server ./apps/gateway/cmd/server/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

# Copy the config file from the host into the container
# Source Path (Host): ./configs/dev.yaml
# Destination Path (Container): /etc/app/config.yaml (Must be the path your Go app expects)
COPY configs/dev.yaml /etc/app/config.yaml

EXPOSE 8080

CMD ["/server"]
