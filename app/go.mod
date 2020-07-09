module github.com/carlosjgp/observability-demo-app

go 1.14

require (
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.4.2
	github.com/slok/go-http-metrics v0.8.0
	go.opentelemetry.io/otel v0.7.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.7.0
	google.golang.org/grpc v1.30.0
)
