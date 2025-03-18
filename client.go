package main

import (
	"fmt"
	"google.golang.org/grpc"
	"micro/grpc_resolver"
	"micro/register"
	"time"
)

type Client struct {
	r        register.Register
	insecure bool
}

type ClientOpt func(*Client)

func WithClientRegister(r register.Register) ClientOpt {
	return func(c *Client) {
		c.r = r
	}
}

func WithInSecure() ClientOpt {
	return func(c *Client) {
		c.insecure = true
	}
}

func NewClient(opts ...ClientOpt) *Client {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Dial(serviceName string, timeout time.Duration) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if c.r != nil {
		rb := grpc_resolver.NewResolverBuilder(c.r, timeout)
		opts = append(opts, grpc.WithResolvers(rb))
	}

	return grpc.NewClient(fmt.Sprintf("passthrough:///%s", serviceName), opts...)
}
