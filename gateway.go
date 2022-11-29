package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	XRealIP    = "X-Real-IP"
	XRequestID = "X-Request-ID"
)

const (
	startBufferSize    = 2048
	grpcConnectTimeOut = time.Second * 10
	shutdownTimeout    = time.Second * 5
)

type MiddlewareFn = func(http.Handler) http.Handler
type GRPCMiddlewareFn = func(desc Description) MiddlewareFn

type Gateway struct {
	http              *http.Server
	mapper            *Mapper
	buffers           *bufferPool
	services          map[string]*GRPCServiceRegistry
	grpcConn          map[string]*grpc.ClientConn
	discover          Discover
	defaultHandler    http.Handler
	grpcDialOptions   []grpc.DialOption
	routerMiddlewares []MiddlewareFn
	grpcMiddlewares   []GRPCMiddlewareFn
	grpcDialTimeout   time.Duration
	shutdownTimeout   time.Duration
}

func (g *Gateway) Run(ctx context.Context) error {
	var err error

	r := chi.NewRouter()
	g.http.Handler = r

	r.Use(middleware.Recoverer)
	r.Use(g.routerMiddlewares...)

	if err = g.setHTTP(r); err != nil {
		return err
	}
	if err = g.setGRPC(ctx, r); err != nil {
		return err
	}
	r.NotFound(g.defaultRoute)

	g.routerMiddlewares = nil
	g.grpcMiddlewares = nil
	g.grpcDialOptions = nil

	return g.http.ListenAndServe()
}

func (g *Gateway) AddGRPCServiceRegistry(service string, registry *GRPCServiceRegistry) {
	g.services[service] = registry
}

func (g *Gateway) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, g.shutdownTimeout)
	defer cancel()

	if err := g.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server %w", err)
	}
	for svc, conn := range g.grpcConn {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("serivce %s: %w", svc, err)
		}
	}
	return nil
}

func (g *Gateway) defaultRoute(w http.ResponseWriter, r *http.Request) {
	if _, ok := g.mapper.peekDefault(); ok {
		if g.defaultHandler == nil {
			fmt.Printf("default route not defined \n")
			if err := httpWriteError(w, ErrInternalEmptyDefaultHandler); err != nil {
				log.Printf("default route write response error: %s \n", err)
			}
			return
		}
		g.defaultHandler.ServeHTTP(w, r)
		return
	}
	fmt.Printf("default route not defined \n")
	if err := httpWriteError(w, ErrNotFound); err != nil {
		log.Printf("default route write response error: %s \n", err)
	}
}

func (g *Gateway) errorHandler(service, address, version string) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("HTTP error handler: %s [%s:@%s]: %s", service, address, version, err)
		if err = httpWriteError(w, ErrInternalServer); err != nil {
			log.Printf("%s [%s:@%s] write response error: %s \n", service, address, version, err)
		}
	}
}

func (g *Gateway) setHTTP(router chi.Router) error {
	for _, svc := range g.mapper.peekServices() {
		if svc.transport == GRPC {
			continue
		}

		addr, err := url.Parse(svc.address)
		if err != nil {
			return fmt.Errorf("parse svc %s on %s %w", svc.name, svc.address, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(addr)
		proxy.BufferPool = g.buffers
		proxy.ErrorHandler = g.errorHandler(svc.name, svc.address, "default")

		versionProxies := make(map[string]*httputil.ReverseProxy, len(svc.versions))
		for _, version := range svc.versions {
			addr, err = url.Parse(version.URL)
			if err != nil {
				return fmt.Errorf("failed to parse url for service: %s, version: %s, url: %s, err: %w",
					svc.name, version.Version, version.URL, err,
				)
			}
			versionProxy := httputil.NewSingleHostReverseProxy(addr)
			versionProxy.BufferPool = g.buffers
			versionProxy.ErrorHandler = g.errorHandler(svc.name, version.URL, version.Version)
			versionProxies[version.Version] = versionProxy
		}

		handler := newHTTPHandler(svc, proxy, versionProxies)
		for _, endpoint := range svc.endpoints {
			router.Handle(endpoint, handler)
		}
		if def, ok := g.mapper.peekDefault(); ok {
			if def.name == svc.name {
				g.defaultHandler = handler
			}
		}
	}
	return nil
}

func (g *Gateway) setGRPC(ctx context.Context, r *chi.Mux) error {
	var svc *description
	for _, svc = range g.mapper.peekServices() {
		if err := g.regGRPC(ctx, svc, r); err != nil {
			return err
		}
	}
	return nil
}

func (g *Gateway) lookupServiceRegistry(service string) (*GRPCServiceRegistry, bool) {
	handler, ok := g.services[service]
	return handler, ok
}

func (g *Gateway) regGRPC(ctx context.Context, svc *description, r *chi.Mux) error {
	if svc.transport == REST {
		return nil
	}
	if svc.prefix == "" {
		return fmt.Errorf("%w, service: `%s`", ErrEmptyGrpcServicePrefix, svc.name)
	}

	desc, ok := g.lookupServiceRegistry(svc.name)
	if !ok {
		return ErrEmptyGrpcServiceDidNotRegistered
	}

	var dest string
	addr := strings.TrimPrefix(svc.address, "grpc://")
	if dest, ok = g.discover(addr); ok {
		addr = dest
	}

	timeOutCtx, cancel := context.WithTimeout(ctx, g.grpcDialTimeout)
	conn, err := grpc.DialContext(timeOutCtx, addr, g.grpcDialOptions...)
	cancel()

	if err != nil {
		return fmt.Errorf("create connection to service %s on %s: %w", svc.name, svc.address, err)
	}

	marshaller := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: false,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customMatcher),
		runtime.WithErrorHandler(g.grpcGatewayErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller),
	)

	if errMuxRegister := desc.register(ctx, mux, conn); errMuxRegister != nil {
		return fmt.Errorf("try register service %s on %s: %w", svc.name, svc.address, errMuxRegister)
	}
	g.grpcConn[svc.name] = conn

	r.Route(svc.prefix, g.grpcRouteFunc(mux, svc))
	return nil
}

func (g *Gateway) grpcGatewayErrorHandler(
	_ context.Context,
	_ *runtime.ServeMux,
	_ runtime.Marshaler,
	w http.ResponseWriter,
	_ *http.Request,
	err error,
) {
	log.Printf("GRPC error handler: %v\n", err)
	if err = httpWriteError(w, ErrInternalServer); err != nil {
		log.Printf("grpc write response error: %s \n", err)
	}
}

func (g *Gateway) grpcRouteFunc(mux http.Handler, svc *description) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				if svc.prefix != "" {
					r.URL.Path = strings.Replace(r.URL.Path, svc.prefix, "", 1)
				}
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(fn)
		})
		for _, md := range g.grpcMiddlewares {
			r.Use(md(svc))
		}
		r.Handle("/", mux)
		r.NotFound(mux.ServeHTTP)
	}
}

func New(options ...Option) (*Gateway, error) {
	g := &Gateway{
		mapper:          NewMapper(),
		buffers:         newBufferPool(startBufferSize),
		services:        map[string]*GRPCServiceRegistry{},
		grpcConn:        make(map[string]*grpc.ClientConn),
		discover:        defaultDiscover,
		grpcDialTimeout: grpcConnectTimeOut,
		shutdownTimeout: shutdownTimeout,
	}
	for _, option := range options {
		option(g)
	}
	return g, nil
}

func customMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case strings.ToLower(XRealIP):
		return key, true
	case strings.ToLower(XRequestID):
		return XRequestID, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func defaultDiscover(service string) (string, bool) {
	return service, true
}
