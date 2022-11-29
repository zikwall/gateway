// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.2
// source: some.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SomeClient is the client API for Some service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SomeClient interface {
	HandlerOne(ctx context.Context, in *HandlerOneRequest, opts ...grpc.CallOption) (*HandlerOneResponse, error)
	HandlerTwo(ctx context.Context, in *HandlerTwoRequest, opts ...grpc.CallOption) (*HandlerTwoResponse, error)
}

type someClient struct {
	cc grpc.ClientConnInterface
}

func NewSomeClient(cc grpc.ClientConnInterface) SomeClient {
	return &someClient{cc}
}

func (c *someClient) HandlerOne(ctx context.Context, in *HandlerOneRequest, opts ...grpc.CallOption) (*HandlerOneResponse, error) {
	out := new(HandlerOneResponse)
	err := c.cc.Invoke(ctx, "/Some.V1.Some/HandlerOne", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *someClient) HandlerTwo(ctx context.Context, in *HandlerTwoRequest, opts ...grpc.CallOption) (*HandlerTwoResponse, error) {
	out := new(HandlerTwoResponse)
	err := c.cc.Invoke(ctx, "/Some.V1.Some/HandlerTwo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SomeServer is the server API for Some service.
// All implementations must embed UnimplementedSomeServer
// for forward compatibility
type SomeServer interface {
	HandlerOne(context.Context, *HandlerOneRequest) (*HandlerOneResponse, error)
	HandlerTwo(context.Context, *HandlerTwoRequest) (*HandlerTwoResponse, error)
	mustEmbedUnimplementedSomeServer()
}

// UnimplementedSomeServer must be embedded to have forward compatible implementations.
type UnimplementedSomeServer struct {
}

func (UnimplementedSomeServer) HandlerOne(context.Context, *HandlerOneRequest) (*HandlerOneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandlerOne not implemented")
}
func (UnimplementedSomeServer) HandlerTwo(context.Context, *HandlerTwoRequest) (*HandlerTwoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandlerTwo not implemented")
}
func (UnimplementedSomeServer) mustEmbedUnimplementedSomeServer() {}

// UnsafeSomeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SomeServer will
// result in compilation errors.
type UnsafeSomeServer interface {
	mustEmbedUnimplementedSomeServer()
}

func RegisterSomeServer(s grpc.ServiceRegistrar, srv SomeServer) {
	s.RegisterService(&Some_ServiceDesc, srv)
}

func _Some_HandlerOne_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandlerOneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SomeServer).HandlerOne(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Some.V1.Some/HandlerOne",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SomeServer).HandlerOne(ctx, req.(*HandlerOneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Some_HandlerTwo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandlerTwoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SomeServer).HandlerTwo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Some.V1.Some/HandlerTwo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SomeServer).HandlerTwo(ctx, req.(*HandlerTwoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Some_ServiceDesc is the grpc.ServiceDesc for Some service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Some_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Some.V1.Some",
	HandlerType: (*SomeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandlerOne",
			Handler:    _Some_HandlerOne_Handler,
		},
		{
			MethodName: "HandlerTwo",
			Handler:    _Some_HandlerTwo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "some.proto",
}
