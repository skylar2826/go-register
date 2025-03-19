package register

import (
	"context"
	"io"
)

type Register interface {
	Register(ctx context.Context, si ServiceInstance) error
	Unregister(ctx context.Context, si ServiceInstance) error
	ListServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(serviceName string) <-chan Event
	io.Closer
}

type ServiceInstance struct {
	Name    string
	Address string
	Weight  uint32
	Group   string
}

type Event struct {
}
