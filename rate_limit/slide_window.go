package rate_limit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindow struct {
	queue *list.List
	// 最大传输速率
	rate int64
	// 窗口大小
	interval int64
	mu       sync.Mutex
}

func NewSlideWindow(rate int64, interval time.Duration) *SlideWindow {
	return &SlideWindow{
		queue:    list.New(),
		rate:     rate,
		interval: int64(interval),
	}
}

func (w *SlideWindow) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		now := time.Now().UnixNano()
		// 快路径
		if int64(w.queue.Len()) < w.rate {
			w.mu.Lock()
			w.queue.PushBack(now)
			w.mu.Unlock()
			resp, err := handler(ctx, req)
			return resp, err
		}

		// 慢路径
		boundary := now - w.interval
		item := w.queue.Front()

		w.mu.Lock()
		for item.Value.(int64) < boundary {
			w.queue.Remove(item)
			item = w.queue.Front()
		}

		if int64(w.queue.Len()) >= w.rate {
			return nil, errors.New("达到处理能力上限了")
		}

		w.queue.PushBack(now)
		w.mu.Unlock()

		resp, err := handler(ctx, req)
		return resp, err
	}
}
