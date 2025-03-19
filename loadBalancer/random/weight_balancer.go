package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

// 加权随机

type WeightBalancerBuilder struct {
}

func (b *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]WeightConn, 0, len(info.ReadySCs))
	var totalWeight uint32
	for c, sc := range info.ReadySCs {
		weight := sc.Address.Attributes.Value("weight").(uint32)
		totalWeight += weight
		connections = append(connections, WeightConn{
			c:      c,
			weight: weight,
		})
	}

	return &WightBalancer{connections: connections, totalWeight: totalWeight}
}

type WightBalancer struct {
	connections []WeightConn
	totalWeight uint32
}

func (b *WightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	weight := rand.Intn(int(b.totalWeight))
	var curConn WeightConn
	for _, conn := range b.connections {
		if int(conn.weight) >= weight {
			return balancer.PickResult{
				SubConn: curConn.c,
				Done: func(info balancer.DoneInfo) {

				},
			}, nil
		}
	}
	return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
}

type WeightConn struct {
	c      balancer.SubConn
	weight uint32
}
