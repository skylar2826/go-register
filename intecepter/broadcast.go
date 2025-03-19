package intecepter

import (
	"context"
	"google.golang.org/grpc"
	"micro/register"
	"reflect"
	"sync"
)

type BroadcastBuilder struct {
	r           register.Register
	serviceName string
	opts        []grpc.DialOption
}

func NewBroadcastBuilder(r register.Register, serverName string, opts ...grpc.DialOption) *BroadcastBuilder {
	return &BroadcastBuilder{
		r:           r,
		serviceName: serverName,
		opts:        opts,
	}
}

func (b *BroadcastBuilder) Build() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ch, ok := isBroadcast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		serviceInstances, err := b.r.ListServices(ctx, b.serviceName)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup

		wg.Add(len(serviceInstances))
		typ := reflect.TypeOf(reply).Elem()
		for _, si := range serviceInstances {
			go func() {
				siCC, er := grpc.Dial(si.Address, b.opts...)
				if er != nil {
					ch <- Resp{
						Err: er,
					}
					return
				}

				newReply := reflect.New(typ).Interface()
				er = invoker(ctx, method, req, newReply, siCC, opts...)
				if er != nil {
					ch <- Resp{
						Err: er,
					}
					return
				}

				// 处理所有响应
				select {
				case <-ctx.Done():
					ch <- Resp{
						Err: ctx.Err(),
					}
					return

				case ch <- Resp{
					Data: newReply,
					Err:  er,
				}:
				}

				defer wg.Done()
			}()
		}

		wg.Wait()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

type Broadcast struct {
}

func isBroadcast(ctx context.Context) (chan Resp, bool) {
	val, ok := ctx.Value(Broadcast{}).(chan Resp)
	return val, ok
}

func UseBroadcast(ctx context.Context) (<-chan Resp, context.Context) {
	ch := make(chan Resp)
	return ch, context.WithValue(ctx, Broadcast{}, ch)
}

type Resp struct {
	Data any
	Err  error
}
