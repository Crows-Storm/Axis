package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
)

// rueidisClient Read/write splitting client implementation
type rueidisClient struct {
	readClient  rueidis.Client
	writeClient rueidis.Client

	healthy     atomic.Bool
	healthStop  chan struct{}
	closeOnce   sync.Once
	healthCheck time.Duration
}

// newRueidisClient Create a read/write splitting client (internal factory method)
func newRueidisClient(readClient, writeClient rueidis.Client, opts ...Option) RueidisClient {
	options := defaultCreateOptions()
	for _, opt := range opts {
		opt(options)
	}

	c := &rueidisClient{
		readClient:  readClient,
		writeClient: writeClient,
		healthStop:  make(chan struct{}),
		healthCheck: options.healthInterval,
	}
	c.healthy.Store(true)

	if options.healthCheckEnabled {
		go c.backgroundHealthCheck()
	}

	return c
}

// ---------- interface Implement ----------

func (c *rueidisClient) R() rueidis.Client {
	return c.readClient
}

func (c *rueidisClient) W() rueidis.Client {
	return c.writeClient
}

func (c *rueidisClient) Do(ctx context.Context, opType OpType, cmd rueidis.Completed) rueidis.RedisResult {
	switch opType {
	case Read:
		return c.readClient.Do(ctx, cmd)
	default:
		return c.writeClient.Do(ctx, cmd)
	}
}

func (c *rueidisClient) DoMulti(ctx context.Context, opType OpType, cmds ...rueidis.Completed) []rueidis.RedisResult {
	switch opType {
	case Read:
		return c.readClient.DoMulti(ctx, cmds...)
	default:
		return c.writeClient.DoMulti(ctx, cmds...)
	}
}

func (c *rueidisClient) DoCache(ctx context.Context, cmd rueidis.Cacheable, ttl time.Duration) rueidis.RedisResult {
	return c.readClient.DoCache(ctx, cmd, ttl)
}

func (c *rueidisClient) B() rueidis.Builder {
	return c.readClient.B()
}

func (c *rueidisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pingCmd := c.readClient.B().Ping().Build()

	var readErr, writeErr error

	// Concurrent inspection of read/write clients
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		readErr = c.readClient.Do(ctx, pingCmd).Error()
	}()

	go func() {
		defer wg.Done()
		writeErr = c.writeClient.Do(ctx, c.writeClient.B().Ping().Build()).Error()
	}()

	wg.Wait()

	if readErr != nil || writeErr != nil {
		c.healthy.Store(false)
		return fmt.Errorf("redis health check failed: read=%v, write=%v", readErr, writeErr)
	}

	c.healthy.Store(true)
	return nil
}

func (c *rueidisClient) IsHealthy() bool {
	return c.healthy.Load()
}

func (c *rueidisClient) Close() {
	c.closeOnce.Do(func() {
		close(c.healthStop)
		c.readClient.Close()
		c.writeClient.Close()
		logger.Info("Redis read/write clients closed")
	})
}

// backgroundHealthCheck Regular health checks in the back office
func (c *rueidisClient) backgroundHealthCheck() {
	ticker := time.NewTicker(c.healthCheck)
	defer ticker.Stop()

	var consecutiveFails int

	for {
		select {
		case <-c.healthStop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := c.HealthCheck(ctx); err != nil {
				consecutiveFails++
				logger.WithFields(logrus.Fields{
					"error":             err,
					"consecutive_fails": consecutiveFails,
				}).Error("Redis health check failed")
			} else {
				if consecutiveFails > 0 {
					logger.WithField("previous_fails", consecutiveFails).Info("Redis recovered")
				}
				consecutiveFails = 0
			}
			cancel()
		}
	}
}
