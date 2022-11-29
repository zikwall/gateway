package gateway

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (n *Error) Error() string {
	return fmt.Sprintf("error [status %d, code %d]: %s", n.Code, n.Status, n.Message)
}

var (
	ErrInternalServer = &Error{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
	}
	ErrInternalEmptyDefaultHandler = &Error{
		Status:  http.StatusInternalServerError,
		Message: "Internal error: empty default proxy handler",
	}
	ErrNotFound = &Error{
		Status:  http.StatusNotFound,
		Message: "Not found: route does not exists",
	}
)

var (
	ErrEmptyGrpcServicePrefix           = errors.New("grpc service has empty prefix")
	ErrEmptyGrpcServiceDidNotRegistered = errors.New("grpc service did not registered")
)

var (
	ErrDuplicateEndpoint = errors.New("endpoint was duplicated")
	ErrUnknownTransport  = errors.New("unknown transport")
	ErrNoEndpoints       = errors.New("no endpoints")
)
