package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/go-chi/chi/middleware"
	"github.com/zikwall/gateway"
	"github.com/zikwall/gateway/discovery"

	v12 "github.com/zikwall/gateway/simple/proto/another/v1"
	v1 "github.com/zikwall/gateway/simple/proto/some/v1"
)

// don't take code panics as something bad here, this is a lightweight version of code without proper error handling
// everything is fine.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mapper, err := gateway.NewMapperFromYamlFile("./example.yml")
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:              "0.0.0.0:1338",
		ReadTimeout:       time.Second * 5,
		WriteTimeout:      time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
	}

	gw, err := gateway.New(
		gateway.WithHTTP(server),
		gateway.WithServiceMapper(mapper),
		gateway.WithDiscovery(discovery.NewEnvironment().Lookup),
		gateway.WithGRPCDialTimeout(time.Second*10),
		gateway.WithServices(map[string]*gateway.GRPCServiceRegistry{
			"another": gateway.NewServiceRegistry(v12.RegisterAnotherHandler),
		}),
		gateway.WithGRPCDialOptions(DefaultDialOptions()...),
		gateway.WithRouterMiddleware(
			XRealIP,
			func(h http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					fmt.Println("Hey, I'm a router middleware, my IP is", r.Header.Get(XRealIPHeader))
					h.ServeHTTP(w, r)
				}
				return http.HandlerFunc(fn)
			},
		),
		gateway.WithGRPCMiddleware(
			func(desc gateway.Description) gateway.MiddlewareFn {
				return func(h http.Handler) http.Handler {
					fn := func(w http.ResponseWriter, r *http.Request) {
						fmt.Printf("Hey, I'm a GRPC middleware, I'm calling service in route [%s] %s/%s%s \n",
							r.Method, desc.Address(), desc.Name(), r.URL.Path,
						)
						h.ServeHTTP(w, r)
					}
					return http.HandlerFunc(fn)
				}
			},
		),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = gw.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	gw.AddGRPCServiceRegistry("some", gateway.NewServiceRegistry(v1.RegisterSomeHandler))

	go func() {
		if err = gw.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
}

// XRealIPHeader http header
// nolint:gochecknoglobals // it's OK
var XRealIPHeader = http.CanonicalHeaderKey("X-Real-IP")

func XRealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set(XRealIPHeader, r.RemoteAddr)
		h.ServeHTTP(w, r)
	}
	return middleware.RealIP(http.HandlerFunc(fn))
}

func DefaultKeepaliveClientOptions() keepalive.ClientParameters {
	return keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}
}

func DefaultDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(DefaultKeepaliveClientOptions()),
		grpc.WithUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
		),
	}
}
