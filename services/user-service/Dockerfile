FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /server services/user-service/cmd/server/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

EXPOSE 50051

CMD ["/server"]
