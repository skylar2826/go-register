package least_active

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync/atomic"
)

// 最小活跃数

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]RandomConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, RandomConn{
			c:   c,
			cnt: math.MaxUint32,
		})
	}

	return &Balancer{connections: connections}
}

type Balancer struct {
	connections []RandomConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	minCnt := uint32(math.MaxUint32)
	var curConn RandomConn
	for _, conn := range b.connections {
		if atomic.LoadUint32(&conn.cnt) <= minCnt {
			curConn = conn
			minCnt = conn.cnt
		}
	}

	atomic.AddUint32(&curConn.cnt, 1)
	return balancer.PickResult{
		SubConn: curConn.c,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&curConn.cnt, -1)
		},
	}, nil
}

type RandomConn struct {
	c   balancer.SubConn
	cnt uint32
}
