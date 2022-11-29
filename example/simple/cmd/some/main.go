package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"google.golang.org/grpc"

	v1 "github.com/zikwall/gateway/simple/proto/some/v1"
)

// don't take code panics as something bad here, this is a lightweight version of code without proper error handling
// everything is fine.
func main() {
	server := grpc.NewServer([]grpc.ServerOption{}...)
	v1.RegisterSomeServer(server, newSomeServer())

	defer func() {
		server.Stop()
	}()

	// nolint:gosec // only for local tests
	listener, err := net.Listen("tcp", "0.0.0.0:1337")
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

type someServerImpl struct {
	v1.UnimplementedSomeServer
}

func (s someServerImpl) HandlerOne(_ context.Context, request *v1.HandlerOneRequest) (*v1.HandlerOneResponse, error) {
	return &v1.HandlerOneResponse{
		System:   runtime.GOARCH,
		Os:       runtime.GOOS,
		Hardware: runtime.Version(),
	}, nil
}

func (s someServerImpl) HandlerTwo(_ context.Context, request *v1.HandlerTwoRequest) (*v1.HandlerTwoResponse, error) {
	return &v1.HandlerTwoResponse{
		Code:  int32(request.Id) + 100,
		Error: fmt.Sprintf("oops... something went wrong: %s", request.Code),
	}, nil
}

func newSomeServer() v1.SomeServer {
	return &someServerImpl{}
}
