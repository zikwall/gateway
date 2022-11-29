package gateway

type Transport uint

const (
	REST Transport = iota + 1
	GRPC
)
