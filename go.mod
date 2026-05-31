module space-microservices-v2

go 1.25.7

replace (
	github.com/aamadaminov/space-microservices-v2/gencoords/config => ./gencoords/config
	github.com/aamadaminov/space-microservices-v2/gencoords/config/metrics => ./gencoords/config/metrics
	github.com/aamadaminov/space-microservices-v2/gencoords/config/otel => ./gencoords/config/otel
	github.com/aamadaminov/space-microservices-v2/gencoords/monitoring => ./gencoords/monitoring
	github.com/aamadaminov/space-microservices-v2/gencoords/telemetry => ./gencoords/telemetry
	github.com/aamadaminov/space-microservices-v2/pkg/gen/proto/gencoords/v1/gencoordsv1 => ./pkg/gen/proto/gencoords/v1
	github.com/aamadaminov/space-microservices-v2/pkg/metrics => ./pkg/metrics
	github.com/aamadaminov/space-microservices-v2/pkg/otel => ./pkg/otel
)

require (
	github.com/aamadaminov/space-microservices-v2/gencoords/config v0.0.0-00010101000000-000000000000
	github.com/aamadaminov/space-microservices-v2/gencoords/monitoring v0.0.0-00010101000000-000000000000
	github.com/aamadaminov/space-microservices-v2/gencoords/telemetry v0.0.0-00010101000000-000000000000
	github.com/aamadaminov/space-microservices-v2/pkg/gen/proto/gencoords/v1/gencoordsv1 v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.69.0
	google.golang.org/grpc v1.81.1
)

require (
	github.com/aamadaminov/space-microservices-v2/gencoords/config/metrics v0.0.0-00010101000000-000000000000 // indirect
	github.com/aamadaminov/space-microservices-v2/gencoords/config/otel v0.0.0-00010101000000-000000000000 // indirect
	github.com/aamadaminov/space-microservices-v2/pkg/metrics v0.0.0-00010101000000-000000000000 // indirect
	github.com/aamadaminov/space-microservices-v2/pkg/otel v0.0.0-00010101000000-000000000000 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.66.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
