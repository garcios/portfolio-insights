# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy local dependencies (user-service)
# We copy the entire directory to ensure we have go.mod and source code for the replace directive
COPY services/user-service ./services/user-service

# Copy application go.mod and go.sum
COPY apps/login-consent-provider/go.mod apps/login-consent-provider/go.sum ./apps/login-consent-provider/

# Set working directory to the application
WORKDIR /app/apps/login-consent-provider

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY apps/login-consent-provider .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o login-consent-provider ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/apps/login-consent-provider/login-consent-provider .

# Copy templates
COPY --from=builder /app/apps/login-consent-provider/templates ./templates

# Expose port
EXPOSE 3002

# Run the application
CMD ["./login-consent-provider"]
