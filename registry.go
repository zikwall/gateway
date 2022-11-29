package gateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type GRPCServiceHandler func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

type GRPCServiceRegistry struct {
	register GRPCServiceHandler
}

func NewServiceRegistry(handler GRPCServiceHandler) *GRPCServiceRegistry {
	return &GRPCServiceRegistry{handler}
}
