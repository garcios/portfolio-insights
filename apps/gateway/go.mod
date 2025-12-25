module github.com/garcios/portfolio-insights/apps/gateway

go 1.24.0

require (
	github.com/99designs/gqlgen v0.17.45
	github.com/garcios/portfolio-insights/pkg v0.0.0
	github.com/garcios/portfolio-insights/services/marketdata-service v0.0.0
	github.com/garcios/portfolio-insights/services/portfolio-service v0.0.0-00010101000000-000000000000
	github.com/garcios/portfolio-insights/services/transaction-service v0.0.0-00010101000000-000000000000
	github.com/garcios/portfolio-insights/services/user-service v0.0.0-00010101000000-000000000000
	github.com/graph-gophers/dataloader/v7 v7.1.2
	github.com/lestrrat-go/jwx/v2 v2.0.21
	github.com/prometheus/client_golang v1.23.2
	github.com/spf13/viper v1.21.0
	github.com/vektah/gqlparser/v2 v2.5.11
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.5 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sosodev/duration v1.2.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
)

replace github.com/garcios/portfolio-insights/pkg => ../../pkg

replace github.com/garcios/portfolio-insights/services/portfolio-service => ../../services/portfolio-service

replace github.com/garcios/portfolio-insights/services/transaction-service => ../../services/transaction-service

replace github.com/garcios/portfolio-insights/services/user-service => ../../services/user-service

replace github.com/garcios/portfolio-insights/services/marketdata-service => ../../services/marketdata-service
