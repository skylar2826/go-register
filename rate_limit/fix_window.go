package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindow struct {
	// 窗口起始位置
	timestamp int64
	// 窗口大小
	interval int64
	// 最大速率
	rate int64
	// 当前速率
	cnt int64
}

func NewFixWindow(interval time.Duration, rate int64) *FixWindow {
	return &FixWindow{
		timestamp: time.Now().UnixNano(),
		interval:  int64(interval),
		rate:      rate,
	}
}

func (f *FixWindow) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		now := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&f.timestamp)
		cnt := atomic.LoadInt64(&f.cnt)
		if timestamp+f.interval < now {
			// 重置窗口
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, now) {
				atomic.StoreInt64(&cnt, 0)
			}
		}

		atomic.AddInt64(&f.cnt, 1)
		if cnt >= f.rate {
			// 达到最大速率需要限流
			return nil, errors.New("达到能力上限")
		}

		resp, err = handler(ctx, req)
		atomic.AddInt64(&f.cnt, -1)
		return resp, err
	}
}
