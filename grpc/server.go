package grpc

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"micro/register"
	"net"
	"time"
)

type Server struct {
	r register.Register
	*grpc.Server
	timeout time.Duration
	name    string
}

type ServerOpt func(server *Server)

func WithServerRegister(r register.Register, timeout time.Duration) ServerOpt {
	return func(c *Server) {
		c.r = r
		c.timeout = timeout
	}
}

type GrpcServerConfig struct {
	Interceptor grpc.UnaryServerInterceptor
}

func NewServer(name string, config GrpcServerConfig, opts ...ServerOpt) *Server {
	grpcServerOpts := make([]grpc.ServerOption, 0, 1)
	if config.Interceptor != nil {
		grpcServerOpts = append(grpcServerOpts, grpc.UnaryInterceptor(config.Interceptor))
	}

	s := &Server{
		Server: grpc.NewServer(grpcServerOpts...),
		name:   name,
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

type ServerConfig struct {
	Weight uint32
	Group  string
}

// Start 用户调用start时，即确认服务已启动成功可以注册
func (s *Server) Start(addr string, config *ServerConfig) error {
	if s.r != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		err := s.r.Register(ctx, register.ServiceInstance{
			Name:    s.name,
			Address: addr,
			Weight:  config.Weight,
			Group:   config.Group,
		})
		if err != nil {
			cancel()
			return err
		}
		cancel()
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	err = s.Serve(l)
	return err
}

func (s *Server) Stop() error {
	if s.r != nil {
		err := s.r.Close()
		if err != nil {
			log.Println("Error closing register server")
		}
	}
	s.GracefulStop()
	return nil
}
