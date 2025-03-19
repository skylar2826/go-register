package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"micro/register"
	"sync"
)

type Register struct {
	client  *clientv3.Client
	cancels []func()
	mu      sync.Mutex
	close   chan struct{}
	sess    *concurrency.Session
}

func NewRegister(client *clientv3.Client) (*Register, error) {
	sess, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}
	return &Register{
		client: client,
		sess:   sess,
	}, nil
}

func (r *Register) Register(ctx context.Context, si register.ServiceInstance) error {
	bs, err := json.Marshal(si)
	if err != nil {
		return err
	}
	key := getInstanceKey(si)
	_, err = r.client.Put(ctx, key, string(bs), clientv3.WithLease(r.sess.Lease()))
	return err
}

func (r *Register) Unregister(ctx context.Context, si register.ServiceInstance) error {
	_, err := r.client.Delete(ctx, getInstanceKey(si))
	return err
}

func (r *Register) ListServices(ctx context.Context, serviceName string) ([]register.ServiceInstance, error) {
	resp, err := r.client.Get(ctx, getServiceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	services := make([]register.ServiceInstance, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		err = json.Unmarshal(kv.Value, &services[i])
		if err != nil {
			continue
		}
	}

	return services, nil
}

func (r *Register) Subscribe(serviceName string) <-chan register.Event {
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mu.Unlock()
	eventChan := r.client.Watch(ctx, getServiceKey(serviceName), clientv3.WithPrefix())
	res := make(chan register.Event)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-eventChan:
				if e.Canceled || e.Err() != nil {
					continue
				}
				for range e.Events {
					res <- register.Event{}
				}
			}
		}
	}()
	return res
}

func (r *Register) Close() error {
	r.mu.Lock()
	cancels := r.cancels
	cancels = nil
	r.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}

	close(r.close)
	return nil
}

func getServiceKey(serviceName string) string {
	return fmt.Sprintf("micro/%s", serviceName)
}

func getInstanceKey(si register.ServiceInstance) string {
	return fmt.Sprintf("micro/%s/%s", si.Name, si.Address)
}

var _ register.Register = &Register{}
