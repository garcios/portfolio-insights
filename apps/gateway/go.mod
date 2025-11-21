module github.com/garcios/portfolio-insights/apps/gateway

go 1.24

require (
	github.com/garcios/portfolio-insights/pkg v0.0.0
	github.com/99designs/gqlgen v0.17.45
	github.com/vektah/gqlparser/v2 v2.5.11
)

replace github.com/garcios/portfolio-insights/pkg => ../../pkg
