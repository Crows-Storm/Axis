package kafka

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// RetryStrategy 指数退避重试
type RetryStrategy struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

func NewRetryStrategy(maxRetries int, baseDelay, maxDelay time.Duration) *RetryStrategy {
	return &RetryStrategy{
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
	}
}

// Do 执行函数，失败时按指数退避重试
func (r *RetryStrategy) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	var err error
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		err = fn(ctx)
		if err == nil {
			return nil
		}

		if attempt == r.maxRetries {
			break
		}

		// 检查 context 是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		delay := r.calculateDelay(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

func (r *RetryStrategy) calculateDelay(attempt int) time.Duration {
	// 指数退避 + 抖动
	delay := float64(r.baseDelay) * math.Pow(2, float64(attempt))
	jitter := rand.Float64() * float64(r.baseDelay) // 0~baseDelay 的随机抖动
	delay = delay + jitter

	if delay > float64(r.maxDelay) {
		delay = float64(r.maxDelay)
	}
	return time.Duration(delay)
}
