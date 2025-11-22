FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

# Download dependencies for the marketdata-service module
WORKDIR /app/services/marketdata-service
RUN go mod download

# Build the server
WORKDIR /app
RUN go build -o /server ./services/marketdata-service/cmd/server/main.go

FROM alpine:latest

WORKDIR /
COPY --from=builder /server /server

EXPOSE 50054

CMD ["/server"]
