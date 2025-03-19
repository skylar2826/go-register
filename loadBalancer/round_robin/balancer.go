package round_robin

import (
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"log"
	"sync/atomic"
)

// 轮询

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))

	for c := range info.ReadySCs {
		connections = append(connections, c)
	}

	fmt.Println("sub conn", connections)

	return &Balancer{
		connections: connections,
		idx:         -1,
		len:         int32(len(connections)),
	}
}

type Balancer struct {
	connections []balancer.SubConn
	idx         int32
	len         int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	atomic.AddInt32(&b.idx, 1)
	conn := b.connections[b.idx%b.len]

	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {
			if info.Err != nil {
				log.Println(info.Err)
				return
			}

		},
	}, nil
}
