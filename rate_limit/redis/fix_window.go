package redis

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

//go:embed lua/fix_window.lua
var luaFixWindow string

type FixWindow struct {
	client redis.Cmdable
	// 窗口大小
	interval time.Duration
	// 最大速率
	rate int64

	limitDimension string
}

func NewFixWindow(client redis.Cmdable, interval time.Duration, rate int64, limitDimension string) *FixWindow {
	return &FixWindow{
		client:         client,
		interval:       interval,
		rate:           rate,
		limitDimension: limitDimension,
	}
}

func (f *FixWindow) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var val string
		val, err = f.Limit(ctx)
		if err != nil {
			return nil, err
		}
		if val == "true" {
			// 达到最大速率需要限流
			return nil, errors.New("达到能力上限")
		}
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func (f *FixWindow) Limit(ctx context.Context) (string, error) {
	val, err := f.client.Eval(ctx, luaFixWindow, []string{f.limitDimension}, f.interval.Milliseconds(), f.rate).Result()
	if err != nil {
		return "false", err
	}

	return val.(string), nil
}
