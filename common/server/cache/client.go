package cache

import (
	"context"
	"time"

	"github.com/redis/rueidis"
)

// OpType Operation type enumeration, remove magic string
type OpType int

const (
	Read OpType = iota
	Write
)

// RueidisClient Unified Interface
type RueidisClient interface {
	// R Return Only-Read Client
	R() rueidis.Client

	// W Return Read-Write Client
	W() rueidis.Client

	// Do Convenient method: Automatically select the client to execute a single command based on opType.
	Do(ctx context.Context, opType OpType, cmd rueidis.Completed) rueidis.RedisResult

	// DoMulti Batch execution, all commands use the same type of client.
	DoMulti(ctx context.Context, opType OpType, cmds ...rueidis.Completed) []rueidis.RedisResult

	// DoCache Execute caching commands on a read-only client (client-side caching)
	DoCache(ctx context.Context, cmd rueidis.Cacheable, ttl time.Duration) rueidis.RedisResult

	// B Get the command constructor (shortcut, equivalent to R().B()).
	B() rueidis.Builder

	// HealthCheck Check the connectivity between the two clients for reading and writing.
	HealthCheck(ctx context.Context) error
	// IsHealthy Returns the current health status (non-blocking).
	IsHealthy() bool

	// Close all client connection
	Close()
}
