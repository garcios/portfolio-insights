module github.com/garcios/portfolio-insights/apps/gateway

go 1.24.0

require (
	github.com/99designs/gqlgen v0.17.45
	github.com/garcios/portfolio-insights/pkg v0.0.0
	github.com/garcios/portfolio-insights/services/portfolio-service v0.0.0-00010101000000-000000000000
	github.com/garcios/portfolio-insights/services/transaction-service v0.0.0-00010101000000-000000000000
	github.com/garcios/portfolio-insights/services/user-service v0.0.0-00010101000000-000000000000
	github.com/vektah/gqlparser/v2 v2.5.11
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/sosodev/duration v1.2.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
)

replace github.com/garcios/portfolio-insights/pkg => ../../pkg

replace github.com/garcios/portfolio-insights/services/portfolio-service => ../../services/portfolio-service

replace github.com/garcios/portfolio-insights/services/transaction-service => ../../services/transaction-service

replace github.com/garcios/portfolio-insights/services/user-service => ../../services/user-service
