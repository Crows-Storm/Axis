package redis

import (
	"time"
)

// Option Client create options
type Option func(*createOptions)

type createOptions struct {
	healthInterval     time.Duration
	healthCheckEnabled bool
}

func defaultCreateOptions() *createOptions {
	return &createOptions{
		healthInterval:     30 * time.Second,
		healthCheckEnabled: true,
	}
}

// WithLogger Set logger
func WithLogger() Option {
	return func(o *createOptions) {
	}
}

// WithHealthCheckInterval Set health check intervals
func WithHealthCheckInterval(d time.Duration) Option {
	return func(o *createOptions) {
		o.healthInterval = d
		o.healthCheckEnabled = d > 0
	}
}

func DisableHealthCheck() Option {
	return func(o *createOptions) {
		o.healthCheckEnabled = false
	}
}
