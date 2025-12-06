FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

# Download dependencies for the user-service module
WORKDIR /app/services/user-service
RUN go mod download

# Build the server
WORKDIR /app
RUN go build -o /server ./services/user-service/cmd/server/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

# Copy the config file from the host into the container
# Source Path (Host): ./configs/dev.yaml
# Destination Path (Container): /etc/app/config.yaml
COPY services/user-service/configs/dev.yaml /etc/app/config.yaml

EXPOSE 50051

CMD ["/server"]
