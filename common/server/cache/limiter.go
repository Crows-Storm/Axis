package cache

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/rueidis"
)

const HttpRateLimitRequest = "rate_limit:%s"

type SlidingWindowLimiter struct {
	client     rueidis.Client
	windowSize int64
	limit      int64
}

func NewSlidingWindowLimiter(client rueidis.Client, windowSize int64, limit int64) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client:     client,
		windowSize: windowSize,
		limit:      limit,
	}
}

var slidingWindowScript = rueidis.NewLuaScript(`
    local key = KEYS[1]
    local windowStart = tonumber(ARGV[1])
    local now = tonumber(ARGV[2])
    local limit = tonumber(ARGV[3])
    local ttl = tonumber(ARGV[4])

    redis.call('ZREMRANGEBYSCORE', key, '-inf', windowStart)
    local currentCount = redis.call('ZCARD', key)

    if currentCount < limit then
        local member = now .. ':' .. math.random(1000000)
        redis.call('ZADD', key, now, member)
        redis.call('EXPIRE', key, ttl)
        return 1
    else
        return 0
    end
`)

func (l *SlidingWindowLimiter) AllowRequest(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()
	windowStart := now - l.windowSize
	redisKey := fmt.Sprintf(HttpRateLimitRequest, key)

	result, err := slidingWindowScript.Exec(ctx, l.client, []string{redisKey}, []string{
		strconv.FormatInt(windowStart, 10),
		strconv.FormatInt(now, 10),
		strconv.FormatInt(l.limit, 10),
		strconv.FormatInt(l.windowSize+1, 10),
	}).AsInt64()

	if err != nil {
		return false, fmt.Errorf("sliding window limiter: %w", err)
	}

	return result == 1, nil
}

var limiter *SlidingWindowLimiter

func InitRateLimiter(rdb rueidis.Client, windowSize int64, limit int64) {
	limiter = NewSlidingWindowLimiter(rdb, windowSize, limit)
}

func RateLimitMiddlewareWithKey(keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := keyFunc(c)
		if key == "" {
			key = c.ClientIP()
		}

		allowed, err := limiter.AllowRequest(ctx, key)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded",
				"data":    nil,
			})
			return
		}

		c.Next()
	}
}
