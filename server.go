package main

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

func NewServer(name string, opts ...ServerOpt) *Server {
	s := &Server{
		Server: grpc.NewServer(),
		name:   name,
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start 用户调用start时，即确认服务已启动成功可以注册
func (s *Server) Start(addr string) error {
	if s.r != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		err := s.r.Register(ctx, register.ServiceInstance{
			Name:    s.name,
			Address: addr,
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
