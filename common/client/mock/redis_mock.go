package mock

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/mock"
)

// RueidisClient is the mock implementation of the redis.RueidisClient interface
// Use testify/mock to precisely control the return value of each method
type RueidisClient struct {
	mock.Mock
	readClient  rueidis.Client
	writeClient rueidis.Client
}

func NewRueidisClient(readClient, writeClient rueidis.Client) *RueidisClient {
	return &RueidisClient{
		readClient:  readClient,
		writeClient: writeClient,
	}
}

func (m *RueidisClient) R() rueidis.Client { return m.readClient }
func (m *RueidisClient) W() rueidis.Client { return m.writeClient }
func (m *RueidisClient) B() rueidis.Builder {
	if m.readClient != nil {
		return m.readClient.B()
	}
	return m.writeClient.B()
}
func (m *RueidisClient) IsHealthy() bool                     { return true }
func (m *RueidisClient) Close()                              {}
func (m *RueidisClient) HealthCheck(_ context.Context) error { return nil }

func (m *RueidisClient) Do(ctx context.Context, opType redis.OpType, cmd rueidis.Completed) rueidis.RedisResult {
	args := m.Called(ctx, opType, cmd)
	return args.Get(0).(rueidis.RedisResult)
}

func (m *RueidisClient) DoMulti(ctx context.Context, opType redis.OpType, cmds ...rueidis.Completed) []rueidis.RedisResult {
	args := m.Called(ctx, opType, cmds)
	return args.Get(0).([]rueidis.RedisResult)
}

func (m *RueidisClient) DoCache(ctx context.Context, cmd rueidis.Cacheable, ttl time.Duration) rueidis.RedisResult {
	args := m.Called(ctx, cmd, ttl)
	return args.Get(0).(rueidis.RedisResult)
}
