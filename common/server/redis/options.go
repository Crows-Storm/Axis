package redis

import (
	"time"

	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
)

type Option func(*createOptions)

type createOptions struct {
	logger         *logrus.Entry
	clientOptions  []func(*rueidis.ClientOption)
	healthInterval *time.Duration
}

func defaultCreateOptions() *createOptions {
	return &createOptions{
		logger: logrus.NewEntry(logrus.StandardLogger()),
	}
}

func WithLogger(logger *logrus.Entry) Option {
	return func(o *createOptions) {
		o.logger = logger
	}
}

func WithClientOption(fn func(*rueidis.ClientOption)) Option {
	return func(o *createOptions) {
		o.clientOptions = append(o.clientOptions, fn)
	}
}

func WithHealthCheckInterval(d time.Duration) Option {
	return func(o *createOptions) {
		o.healthInterval = &d
	}
}

func DisableHealthCheck() Option {
	return func(o *createOptions) {
		zero := time.Duration(0)
		o.healthInterval = &zero
	}
}
