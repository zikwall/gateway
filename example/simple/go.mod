module github.com/zikwall/gateway/simple

go 1.19

replace github.com/zikwall/gateway => ../../

require (
	github.com/go-chi/chi v1.5.4
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.14.0
	github.com/zikwall/gateway v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.36.4
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/trace v1.11.1
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6
	google.golang.org/grpc v1.51.0
	google.golang.org/protobuf v1.28.1
)

require (
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.1.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
