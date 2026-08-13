package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
)

// Initialize the read-write splitting Redis client and register it to the global registry.
func Initialize(readCfg config.ReadRedisConfig, writeCfg config.WriteRedisConfig, redisHealthCheckInterval int) error {
	readClient, err := createRawClient(readCfg.Address, readCfg.Password, readCfg.DB, readCfg.DisableCache)
	if err != nil {
		return fmt.Errorf("create read redis client failed: %w", err)
	}

	writeClient, err := createRawClient(writeCfg.Address, writeCfg.Password, writeCfg.DB, writeCfg.DisableCache)
	if err != nil {
		readClient.Close()
		return fmt.Errorf("create write redis client failed: %w", err)
	}

	client := newRueidisClient(
		readClient,
		writeClient,
		WithHealthCheckInterval(time.Duration(redisHealthCheckInterval)*time.Second),
	)

	if err := Register("cache", client); err != nil {
		client.Close()
		return fmt.Errorf("register redis client failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"read_addrs":  readCfg.Address,
		"write_addrs": writeCfg.Address,
		"read_db":     readCfg.DB,
		"write_db":    writeCfg.DB,
	}).Info("Redis read/write client initialized successfully")

	return nil
}

// createRawClient Create the underlying rueidis.Client and verify connectivity.
func createRawClient(addrs []string, password string, db int, disableCache bool) (rueidis.Client, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("redis address is empty")
	}

	opt := rueidis.ClientOption{
		InitAddress:      addrs,
		Password:         password,
		SelectDB:         db,
		DisableCache:     disableCache,
		AlwaysPipelining: true,
		MaxFlushDelay:    20 << 10, // 20KB
		// TODO
		//SendToReplicas: func(cmd rueidis.Completed) bool {
		//	return true
		//},
	}

	rdb, err := rueidis.NewClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create rueidis client failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Do(ctx, rdb.B().Ping().Build()).Error(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return rdb, nil
}

// GetClient Get Global Client
func GetClient() (RueidisClient, error) {
	return Get("cache")
}

// MustGetClient 获取全局客户端，失败则 panic
func MustGetClient() RueidisClient {
	return MustGet("cache")
}
