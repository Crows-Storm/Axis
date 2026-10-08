package rbac

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/redis/rueidis"
)

type Redis struct {
	client cache.RueidisClient
}

func NewRedis(client cache.RueidisClient) *Redis {
	if client == nil {
		return nil
	}
	return &Redis{client: client}
}

func (r *Redis) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.client.Do(ctx, cache.Read, r.client.B().Get().Key(key).Build()).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("redis get %s: %w", key, err)
	}
	return val, true, nil
}

func (r *Redis) Set(ctx context.Context, key, value string) error {
	if err := r.client.Do(ctx, cache.Write, r.client.B().Set().Key(key).Value(value).Build()).Error(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Del(ctx context.Context, key string) error {
	if err := r.client.Do(ctx, cache.Write, r.client.B().Del().Key(key).Build()).Error(); err != nil {
		return fmt.Errorf("redis del %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Do(ctx, cache.Read, r.client.B().Exists().Key(key).Build()).AsInt64()
	if err != nil {
		return false, fmt.Errorf("redis exists %s: %w", key, err)
	}
	return n > 0, nil
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	vals, err := r.client.Do(ctx, cache.Read, r.client.B().Smembers().Key(key).Build()).AsStrSlice()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis smembers %s: %w", key, err)
	}
	return vals, nil
}

func (r *Redis) SAdd(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	if err := r.client.Do(ctx, cache.Write, r.client.B().Sadd().Key(key).Member(members...).Build()).Error(); err != nil {
		return fmt.Errorf("redis sadd %s: %w", key, err)
	}
	return nil
}

func (r *Redis) HGet(ctx context.Context, key, field string) (string, bool, error) {
	val, err := r.client.Do(ctx, cache.Read, r.client.B().Hget().Key(key).Field(field).Build()).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("redis hget %s %s: %w", key, field, err)
	}
	return val, true, nil
}

func (r *Redis) HSet(ctx context.Context, key string, fields map[string]string) error {
	if len(fields) == 0 {
		return nil
	}
	cmd := r.client.B().Hset().Key(key).FieldValue()
	for field, value := range fields {
		cmd = cmd.FieldValue(field, value)
	}
	if err := r.client.Do(ctx, cache.Write, cmd.Build()).Error(); err != nil {
		return fmt.Errorf("redis hset %s: %w", key, err)
	}
	return nil
}
