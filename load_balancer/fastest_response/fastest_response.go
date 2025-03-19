package fastest_response

import (
	"encoding/json"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"log"
	"net/http"
	"runtime"
	"time"
)

// 最快响应

type BalancerBuilder struct {
	Duration    time.Duration
	connections []fastRespConn
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]fastRespConn, 0, len(info.ReadySCs))
	for c, sc := range info.ReadySCs {
		connections = append(connections, fastRespConn{
			c:        c,
			respTime: time.Millisecond * 100,
			Addr:     sc.Address.Addr,
		})
	}

	res := &Balancer{connections: b.connections}

	closer := make(chan struct{})
	t := time.NewTicker(b.Duration)

	runtime.SetFinalizer(res, func() {
		closer <- struct{}{}
	})

	go func() {
		select {
		case <-t.C:
			res.updateRespTime()
		case <-closer:
			return
		}
	}()

	b.connections = connections
	return res
}

type Response struct {
	Data map[string]time.Duration
}

type Balancer struct {
	connections []fastRespConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var curConn *fastRespConn
	for _, conn := range b.connections {
		if curConn == nil || conn.respTime < curConn.respTime {
			curConn = &conn
		}
	}

	if curConn == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	return balancer.PickResult{
		SubConn: curConn.c,
		Done: func(info balancer.DoneInfo) {
		},
	}, nil
}

func (b *Balancer) updateRespTime() {
	resp, err := http.Get("http://premethes")
	if err != nil {
		log.Printf("Error fetching premethes: %v", err)
		return
	}

	defer resp.Body.Close()

	response := &Response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Printf("Error fetching premethes: %v", err)
		return
	}

	for key, duration := range response.Data {
		for _, conn := range b.connections {
			if conn.Addr == key {
				conn.respTime = duration
			}
		}
	}
}

type fastRespConn struct {
	c        balancer.SubConn
	respTime time.Duration
	Addr     string
}
