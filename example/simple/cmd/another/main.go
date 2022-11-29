package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	v1 "github.com/zikwall/gateway/simple/proto/another/v1"
)

// don't take code panics as something bad here, this is a lightweight version of code without proper error handling
// everything is fine.
func main() {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	)
	v1.RegisterAnotherServer(server, newAnotherServer(otel.Tracer("gateway")))
	httpServer := &http.Server{
		Addr:              ":8082",
		ReadTimeout:       time.Second * 5,
		WriteTimeout:      time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
	}

	defer func() {
		server.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	go func() {
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))

		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// nolint:gosec // only for local tests
	listener, err := net.Listen("tcp", "0.0.0.0:1339")
	if err != nil {
		panic(err)
	}
	if err = server.Serve(listener); err != nil {
		panic(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
}

type anotherServerImpl struct {
	v1.UnimplementedAnotherServer
	tracer trace.Tracer
}

func (a *anotherServerImpl) HandlerOne(ctx context.Context, req *v1.HandlerOneRequest) (*v1.HandlerOneResponse, error) {
	ctx, span := a.tracer.Start(ctx, "handler_one")
	defer span.End()

	span.SetAttributes(
		attribute.String("code", req.Code),
		attribute.Int64("lang", req.Lang),
	)

	return &v1.HandlerOneResponse{
		Code:    req.Code,
		Message: fmt.Sprintf("message: %d", req.Lang),
		Title:   req.String(),
	}, nil
}

func (a *anotherServerImpl) HandlerTwo(ctx context.Context, req *v1.HandlerTwoRequest) (*v1.HandlerTwoResponse, error) {
	ctx, span := a.tracer.Start(ctx, "handle_two")
	defer span.End()

	span.SetAttributes(
		attribute.String("language", req.Language),
		attribute.Int64("id", req.LanguageId),
	)

	id, language := a.someFn(ctx, req.LanguageId, req.Language)
	return &v1.HandlerTwoResponse{
		Id:        id,
		ErrorCode: language,
	}, nil
}

func (a *anotherServerImpl) someFn(ctx context.Context, id int64, language string) (int32, string) {
	ctx, span := a.tracer.Start(ctx, "some_fn")
	defer span.End()

	return int32(id) + 100, fmt.Sprintf("language is wrong: %s", language)
}

func newAnotherServer(tracer trace.Tracer) v1.AnotherServer {
	return &anotherServerImpl{tracer: tracer}
}
