FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /server apps/gateway/cmd/server/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

EXPOSE 8080

CMD ["/server"]
