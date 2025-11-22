module github.com/garcios/portfolio-insights/services/transaction-service

go 1.24.0

require (
	github.com/garcios/portfolio-insights/pkg v0.0.0
	github.com/garcios/portfolio-insights/services/marketdata-service v0.0.0
	github.com/garcios/portfolio-insights/services/user-service v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
)

replace github.com/garcios/portfolio-insights/pkg => ../../pkg

replace github.com/garcios/portfolio-insights/services/user-service => ../user-service

replace github.com/garcios/portfolio-insights/services/marketdata-service => ../marketdata-service
