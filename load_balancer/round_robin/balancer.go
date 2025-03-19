package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"log"
	"micro/route_strategy"
	"sync/atomic"
)

// 轮询

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]Conn, 0, len(info.ReadySCs))

	for c, sc := range info.ReadySCs {
		connections = append(connections, Conn{
			c:    c,
			addr: sc.Address,
		})
	}

	return &Balancer{
		connections: connections,
		idx:         -1,
		len:         int32(len(connections)),
		filter:      route_strategy.GroupFilterBuilder{}.Build(),
	}
}

type Balancer struct {
	connections []Conn
	idx         int32
	len         int32
	filter      route_strategy.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	candidates := make([]Conn, 0, len(b.connections))
	for _, conn := range b.connections {
		if b.filter(info, conn.addr) {
			candidates = append(candidates, conn)
		}
	}

	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	atomic.AddInt32(&b.idx, 1)
	conn := candidates[b.idx%int32(len(candidates))]

	return balancer.PickResult{
		SubConn: conn.c,
		Done: func(info balancer.DoneInfo) {
			if info.Err != nil {
				log.Println(info.Err)
				return
			}

		},
	}, nil
}

type Conn struct {
	c    balancer.SubConn
	addr resolver.Address
}
