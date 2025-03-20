package rate_limit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

// 令牌桶算法

type TokenBucket struct {
	tokens chan struct{}
	close  chan struct{}
}

// NewTokenBucket capacity 桶容量， interval 多久发一张令牌
func NewTokenBucket(capacity int, interval time.Duration) *TokenBucket {
	t := &TokenBucket{}
	t.tokens = make(chan struct{}, capacity)
	t.close = make(chan struct{})
	producer := time.NewTicker(interval)

	go func() {
		defer producer.Stop()
		for {
			select {
			case <-producer.C:
				select {
				case t.tokens <- struct{}{}:
				default:
					// 防止没人取走令牌阻塞
				}
			case <-t.close:
				return
			}
		}
	}()
	return t
}

func (t *TokenBucket) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-t.close:
			resp, err = handler(ctx, req)
			return
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.tokens:
			resp, err = handler(ctx, req)
			return
		}
	}
}

func (t *TokenBucket) Close() {
	close(t.close)
}
