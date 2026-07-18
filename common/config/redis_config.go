package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/rueidis"
	"github.com/spf13/viper"
)

var RedisClient rueidis.Client

const (
	RedisHost     = "redis.host"
	RedisDatabase = "redis.database"
	RedisPassword = "redis.password"
	RedisCache    = "redis.cache"
)

type RedisConfig struct {
	Host     []string
	Database int
	Password string
	Cache    bool
}

func NewRedisConnect() error {
	redisConfig := &RedisConfig{
		Host:     viper.GetStringSlice(RedisHost),
		Database: viper.GetInt(RedisDatabase),
		Password: viper.GetString(RedisPassword),
		Cache:    viper.GetBool(RedisCache),
	}

	initAddresses := make([]string, 0, len(redisConfig.Host))
	for _, host := range redisConfig.Host {
		initAddresses = append(initAddresses, host)
	}

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:   initAddresses,
		Password:      redisConfig.Password,
		SelectDB:      redisConfig.Database,
		DisableCache:  redisConfig.Cache,
		MaxFlushDelay: 2 * time.Millisecond,
	})
	if err != nil {
		return fmt.Errorf("[Redis Config] failed to connect redis: %w", err)
	}

	RedisClient = client

	ctx := context.Background()
	err = RedisClient.Do(ctx, RedisClient.B().Ping().Build()).Error()
	if err != nil {
		return fmt.Errorf("[Redis Config] redis ping failed: %w", err)
	}

	log.Println("[Redis Config] ✅ Redis connected successfully")
	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
		log.Println("[Redis Config] ⛓️‍💥 Redis connection closed")
	}
}
