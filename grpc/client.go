package grpc

import (
	"fmt"
	"google.golang.org/grpc"
	"micro/grpc_resolver"
	"micro/register"
	"time"
)

type Client struct {
	r            register.Register
	insecure     bool
	balancerName string
	interceptor  grpc.UnaryClientInterceptor
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

func WithBalancer(name string) ClientOpt {
	return func(c *Client) {
		c.balancerName = name
	}
}

func WithInterceptor(interceptor grpc.UnaryClientInterceptor) ClientOpt {
	return func(c *Client) {
		c.interceptor = interceptor
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
	if c.balancerName != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"`+c.balancerName+`"}`))
	}
	if c.interceptor != nil {
		opts = append(opts, grpc.WithUnaryInterceptor(c.interceptor))
	}

	return grpc.NewClient(fmt.Sprintf("passthrough:///%s", serviceName), opts...)
}
