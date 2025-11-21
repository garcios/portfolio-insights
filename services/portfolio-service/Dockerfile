FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

# Download dependencies for the portfolio-service module
WORKDIR /app/services/portfolio-service
RUN go mod download

# Build the server
WORKDIR /app
RUN go build -o /server ./services/portfolio-service/cmd/server/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

EXPOSE 50052

CMD ["/server"]
