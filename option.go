package gateway

import (
	"net/http"
	"time"

	"google.golang.org/grpc"
)

type Option func(gateway *Gateway)
type Discover func(service string) (string, bool)

func WithHTTP(server *http.Server) Option {
	return func(g *Gateway) {
		g.http = server
	}
}

func WithDiscovery(discover Discover) Option {
	return func(g *Gateway) {
		g.discover = discover
	}
}

func WithServices(services map[string]*GRPCServiceRegistry) Option {
	return func(g *Gateway) {
		g.services = services
	}
}

func WithServiceMapper(mapper *Mapper) Option {
	return func(g *Gateway) {
		g.mapper = mapper
	}
}

func WithGRPCDialTimeout(timeout time.Duration) Option {
	return func(g *Gateway) {
		g.grpcDialTimeout = timeout
	}
}

func WithRouterMiddleware(middlewares ...MiddlewareFn) Option {
	return func(g *Gateway) {
		g.routerMiddlewares = middlewares
	}
}

func WithGRPCMiddleware(middlewares ...GRPCMiddlewareFn) Option {
	return func(g *Gateway) {
		g.grpcMiddlewares = middlewares
	}
}

func WithGRPCDialOptions(options ...grpc.DialOption) Option {
	return func(g *Gateway) {
		g.grpcDialOptions = options
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(g *Gateway) {
		g.shutdownTimeout = timeout
	}
}
