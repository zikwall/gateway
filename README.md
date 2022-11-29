## Gateway: library HTTP over gRPC

An additional way to access our gRPC service with little effort (easier than writing your own code and serializer), 
which can be useful in a number of cases.

### How to use?

0. Create config file

```yaml
services:
  - service: "some"
    url: "grpc://0.0.0.0:1337"
    prefix: "/some"
    auth: false
    endpoints:
     - /v1/public/handler_one
     - /v1/private/handler_two
default:
  service: "Default"
  url: "http://0.0.0.0:80"
```

2. Define proto file, see example [proto](./example/simple/proto/some/v1/some.proto)

**Note**: it is necessary to use `grpc-gateway_*` flags when generating protobuf code, [example](./example/simple/Makefile)

```shell
--grpc-gateway_out=./proto/$@/v1 \
--grpc-gateway_opt=logtostderr=true \
--grpc-gateway_opt=paths=source_relative \
```

```protobuf
syntax = "proto3";

package Some.V1;

// require import
import "google/api/annotations.proto";

service Some {
  rpc HandlerOne (...) returns (...) {
    option (google.api.http) = {
      get: "/v1/public/handler_one"
    };
  };

  rpc HandlerTwo (...) returns (...) {
    option (google.api.http) = {
      post: "/v1/private/handler_two"
      body: "*"
    };
  };
}
```

2. Generate protobuf files and connect to project, then initialize gateway package, [example](./example/simple/cmd/gateway/main.go)

```go
gw, err := gateway.New(
    gateway.WithHTTP(server),
    gateway.WithServiceMapper(mapper), 
    gateway.WithDiscovery(discovery.NewEnvironment().Lookup),
    gateway.WithGRPCDialTimeout(time.Second*10), 
    gateway.WithServices(map[string]*gateway.GRPCServiceRegistry{
        "another": gateway.NewServiceRegistry(v12.RegisterAnotherHandler),
    }),
    gateway.WithGRPCDialOptions([]grpc.DialOption{}...),
    gateway.WithRouterMiddleware(httpMiddlewares...),
    gateway.WithGRPCMiddleware(grpcMiddlewares, ),
)

// register another one service using API
gw.AddGRPCServiceRegistry("some", gateway.NewServiceRegistry(v1.RegisterSomeHandler))

// and run
if err = gw.Run(ctx); err != nil {
    panic(err)
}
```

3. Make request

```shell
$ curl -L http://localhost:1338/some/v1/public/handler_one?code=500&lang=zh

// or POST

$ curl -L -X POST 'http://localhost:1338/some/v1/private/handler_two' -H 'Content-Type: application/json' --data-raw '{}'
```