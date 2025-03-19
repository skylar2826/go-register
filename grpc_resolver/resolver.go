package grpc_resolver

import (
	"context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"micro/register"
	"time"
)

type ResolverBuilder struct {
	r       register.Register
	timeout time.Duration
}

func NewResolverBuilder(r register.Register, timeout time.Duration) *ResolverBuilder {
	return &ResolverBuilder{
		r:       r,
		timeout: timeout,
	}
}

func (r *ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	rl := &Resolver{cc: cc, r: r.r, timeout: r.timeout, target: target, close: make(chan struct{}, 1)}
	go rl.watch()
	rl.ResolveNow(resolver.ResolveNowOptions{})
	return rl, nil
}

func (r *ResolverBuilder) Scheme() string {
	return "passthrough"
}

type Resolver struct {
	cc      resolver.ClientConn
	r       register.Register
	timeout time.Duration
	target  resolver.Target
	close   chan struct{}
}

func (r *Resolver) watch() {
	event := r.r.Subscribe(r.target.Endpoint())
	for {
		select {
		case <-r.close:
			return
		case <-event:
			r.resolve(resolver.ResolveNowOptions{})
		}
	}
}

func (r *Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.resolve(options)
}

func (r *Resolver) resolve(options resolver.ResolveNowOptions) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	si, err := r.r.ListServices(ctx, r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}

	addresses := make([]resolver.Address, 0, len(si))

	for _, s := range si {
		addresses = append(addresses, resolver.Address{
			Addr:       s.Address,
			Attributes: attributes.New("weight", s.Weight).WithValue("group", s.Group),
		})
	}

	err = r.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
	if err != nil {
		r.cc.ReportError(err)
	}
}

func (r *Resolver) Close() {
	close(r.close)
}
