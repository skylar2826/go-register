package rate_limit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucket struct {
	close chan struct{}
	ch    chan struct{}
}

func NewLeakyBucket(interval time.Duration) *LeakyBucket {
	l := &LeakyBucket{
		ch:    make(chan struct{}),
		close: make(chan struct{}),
	}

	timer := time.NewTicker(interval)

	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				select {
				case l.ch <- struct{}{}:
				default: // 防止阻塞
				}
			case <-l.close:
				return
			}
		}

	}()

	return l
}

func (l *LeakyBucket) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-l.close:
			resp, err = handler(ctx, req)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-l.ch:
			resp, err = handler(ctx, req)
		}

		return resp, err
	}
}
