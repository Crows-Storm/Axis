package config

import (
	"time"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Name         string   `mapstructure:"name"`
	Addrs        []string `mapstructure:"addrs"`
	Password     string   `mapstructure:"password"`
	DB           int      `mapstructure:"db"`
	DisableCache bool     `mapstructure:"disable_cache"`

	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`

	MaxFlushDelay time.Duration `mapstructure:"max_flush_delay"`

	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
}

// LoadRedisConfigs 从 viper Loading all redis instances configuration
// YAML Example:
// redis:
//
//	instances:
//	  - name: cache
//	    addrs: ["localhost:6379"] # cache
//	    db: 0
//	    disable_cache: false
//	    health_check_interval: 30s
//	  - name: session
//	    addrs: ["localhost:6379"] # sessions
//	    db: 1
//	    password: "secret"
func LoadRedisConfigs() ([]RedisConfig, error) {
	var configs []RedisConfig
	if err := viper.UnmarshalKey("redis.instances", &configs); err != nil {
		return nil, err
	}

	for i := range configs {
		if configs[i].Name == "" {
			configs[i].Name = "default"
		}
		if configs[i].DialTimeout == 0 {
			configs[i].DialTimeout = 5 * time.Second
		}
		if configs[i].ReadTimeout == 0 {
			configs[i].ReadTimeout = 3 * time.Second
		}
		if configs[i].WriteTimeout == 0 {
			configs[i].WriteTimeout = 3 * time.Second
		}
		if configs[i].MaxFlushDelay == 0 {
			configs[i].MaxFlushDelay = 2 * time.Millisecond
		}
		if configs[i].HealthCheckInterval == 0 {
			configs[i].HealthCheckInterval = 30 * time.Second
		}
	}

	return configs, nil
}
