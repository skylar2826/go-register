package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"log"
	"math"
	"sync"
)

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*WeightBalancerConn, 0, len(info.ReadySCs))
	for c, sc := range info.ReadySCs {
		wbc := &WeightBalancerConn{
			c:               c,
			weight:          sc.Address.Attributes.Value("weight").(uint32),
			curWeight:       sc.Address.Attributes.Value("weight").(uint32),
			efficientWeight: sc.Address.Attributes.Value("weight").(uint32),
		}

		connections = append(connections, wbc)
	}

	return &WeightBalancer{
		connections: connections,
	}
}

type WeightBalancer struct {
	connections []*WeightBalancerConn
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var curConn *WeightBalancerConn

	for _, conn := range w.connections {
		conn.mu.Lock()
		totalWeight += conn.efficientWeight
		if conn.curWeight >= math.MaxUint32 {
			continue
		}
		conn.curWeight += conn.efficientWeight

		if curConn == nil || conn.curWeight >= curConn.weight {
			curConn = conn
		}
		conn.mu.Unlock()
	}

	if curConn == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	curConn.mu.Lock()
	curConn.curWeight -= totalWeight
	curConn.mu.Unlock()

	return balancer.PickResult{
		SubConn: curConn.c,
		Done: func(info balancer.DoneInfo) {
			curConn.mu.Lock()
			// 调整efficientWeight权重
			if info.Err != nil {
				log.Println(info.Err)
				if curConn.efficientWeight > uint32(0) {
					curConn.efficientWeight--
				}
			} else {
				if curConn.efficientWeight < math.MaxUint32 {
					curConn.efficientWeight++
				}
			}
			curConn.mu.Unlock()
		},
	}, nil
}

type WeightBalancerConn struct {
	c               balancer.SubConn
	weight          uint32
	curWeight       uint32
	efficientWeight uint32
	mu              sync.Mutex
}
