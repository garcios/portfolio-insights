FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/transaction-service

RUN go mod download

WORKDIR /app

RUN go build -o /server ./services/transaction-service/cmd/server/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

# Copy the config file from the host into the container
# Source Path (Host): ./configs/dev.yaml
# Destination Path (Container): /etc/app/config.yaml
COPY services/transaction-service/configs/dev.yaml /etc/app/config.yaml

EXPOSE 50053

CMD ["/server"]
