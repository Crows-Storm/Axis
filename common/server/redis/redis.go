package redis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
)

type RueidisClient interface {
	// Do Single Execution of command
	Do(ctx context.Context, cmd rueidis.Completed) rueidis.RedisResult
	// DoMulti Batch execution of commands（pipeline）
	DoMulti(ctx context.Context, cmds ...rueidis.Completed) []rueidis.RedisResult
	// DoCache Execution Cache command（client-side caching）
	DoCache(ctx context.Context, cmd rueidis.Cacheable, ttl time.Duration) rueidis.RedisResult
	// B Get Command Constructor
	B() rueidis.Builder
	// Receive Use Pub/Sub
	Receive(ctx context.Context, subscribe rueidis.Completed, fn func(msg rueidis.PubSubMessage)) error
	// Close Connection
	Close()
	// HealthCheck Check connectivity
	HealthCheck(ctx context.Context) error
	// IsHealthy Return to current health status (non-blocked)
	IsHealthy() bool
	// Name Return Instances name
	Name() string
}

type rueidisClient struct {
	name        string
	rdb         rueidis.Client
	logger      *logrus.Entry
	healthy     atomic.Bool
	healthStop  chan struct{}
	closeOnce   sync.Once
	healthCheck time.Duration // 0 = No backend health checks are conducted
}

//func NewRueidisClient(name string, rdb rueidis.Client, logger *logrus.Entry, healthInterval time.Duration) RueidisClient {
//	c := &rueidisClient{
//		name:        name,
//		rdb:         rdb,
//		logger:      logger.WithField("instance", name),
//		healthStop:  make(chan struct{}),
//		healthCheck: healthInterval,
//	}
//	c.healthy.Store(true)
//
//	if healthInterval > 0 {
//		go c.backgroundHealthCheck()
//	}
//
//	return c
//}

func NewRueidisClient(name string, rdb rueidis.Client, opts ...Option) RueidisClient {
	options := defaultCreateOptions()
	for _, opt := range opts {
		opt(options)
	}

	var healthInterval time.Duration
	if options.healthInterval != nil {
		healthInterval = *options.healthInterval
	}

	c := &rueidisClient{
		name:        name,
		rdb:         rdb,
		logger:      options.logger.WithField("instance", name),
		healthStop:  make(chan struct{}),
		healthCheck: healthInterval,
	}
	c.healthy.Store(true)

	if healthInterval > 0 {
		go c.backgroundHealthCheck()
	}

	return c
}

func (c *rueidisClient) Do(ctx context.Context, cmd rueidis.Completed) rueidis.RedisResult {
	return c.rdb.Do(ctx, cmd)
}

func (c *rueidisClient) DoMulti(ctx context.Context, cmds ...rueidis.Completed) []rueidis.RedisResult {
	return c.rdb.DoMulti(ctx, cmds...)
}

func (c *rueidisClient) DoCache(ctx context.Context, cmd rueidis.Cacheable, ttl time.Duration) rueidis.RedisResult {
	return c.rdb.DoCache(ctx, cmd, ttl)
}

func (c *rueidisClient) B() rueidis.Builder {
	return c.rdb.B()
}

func (c *rueidisClient) Receive(ctx context.Context, subscribe rueidis.Completed, fn func(msg rueidis.PubSubMessage)) error {
	return c.rdb.Receive(ctx, subscribe, fn)
}

func (c *rueidisClient) Close() {
	c.closeOnce.Do(func() {
		close(c.healthStop)
		c.rdb.Close()
		c.logger.Info("Redis connection closed")
	})
}

func (c *rueidisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := c.rdb.Do(ctx, c.rdb.B().Ping().Build()).Error()
	if err != nil {
		c.healthy.Store(false)
		return fmt.Errorf("redis[%s] ping failed: %w", c.name, err)
	}
	c.healthy.Store(true)
	return nil
}

func (c *rueidisClient) IsHealthy() bool {
	return c.healthy.Load()
}

func (c *rueidisClient) Name() string {
	return c.name
}

// backgroundHealthCheck Regularly check for active activities, log and update status when failures occur
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
			if err := c.rdb.Do(ctx, c.rdb.B().Ping().Build()).Error(); err != nil {
				consecutiveFails++
				c.healthy.Store(false)
				c.logger.WithFields(logrus.Fields{
					"error":             err,
					"consecutive_fails": consecutiveFails,
				}).Error("Redis health check failed")
			} else {
				if consecutiveFails > 0 {
					c.logger.WithField("previous_fails", consecutiveFails).Info("Redis recovered")
				}
				consecutiveFails = 0
				c.healthy.Store(true)
			}
			cancel()
		}
	}
}

func Initialize(cfgs []config.RedisConfig, logger *logrus.Entry) error {
	return Redis(cfgs, logger)
}

func Redis(cfgs []config.RedisConfig, logger *logrus.Entry) error {
	for _, cfg := range cfgs {
		if err := initOneRedis(cfg, logger); err != nil {
			return fmt.Errorf("init redis[%s]: %w", cfg.Name, err)
		}
	}
	return nil
}

func initOneRedis(cfg config.RedisConfig, logger *logrus.Entry) error {
	opt := rueidis.ClientOption{
		InitAddress:      cfg.Addrs,
		Password:         cfg.Password,
		SelectDB:         cfg.DB,
		DisableCache:     cfg.DisableCache,
		MaxFlushDelay:    cfg.MaxFlushDelay,
		ConnWriteTimeout: cfg.WriteTimeout,
	}

	rdb, err := rueidis.NewClient(opt)
	if err != nil {
		return fmt.Errorf("create rueidis client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Do(ctx, rdb.B().Ping().Build()).Error(); err != nil {
		rdb.Close()
		return fmt.Errorf("ping failed: %w", err)
	}

	c := NewRueidisClient(
		cfg.Name,
		rdb,
		WithLogger(logger),
		WithHealthCheckInterval(cfg.HealthCheckInterval),
		// Can expand, e.g.
		// WithClientOption(func(opt *rueidis.ClientOption) { ... }),
		// DisableHealthCheck(),
	)

	if err := Register(cfg.Name, c); err != nil {
		c.Close()
		return err
	}

	logger.WithFields(logrus.Fields{
		"instance":       cfg.Name,
		"addrs":          cfg.Addrs,
		"db":             cfg.DB,
		"cache_disabled": cfg.DisableCache,
	}).Info("Redis initialized")

	return nil
}
