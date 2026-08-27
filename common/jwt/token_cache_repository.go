package jwt

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

const (
	refreshTokenPrefix = "auth:refresh:%s"   // auth:refresh:{jti} -> userID
	accessTokenPrefix  = "auth:blacklist:%s" // auth:blacklist:{jti} -> "1"
)

// TokenCacheRepository Basic Redis (rueidis) the TokenRepository implement
type TokenCacheRepository struct {
	CacheClient rueidis.Client
}

func NewTokenCacheRepository(client rueidis.Client) *TokenCacheRepository {
	return &TokenCacheRepository{
		CacheClient: client,
	}
}

// StoreRefreshToken store refreshToken the JTI → userID mapping，Automatic expiration with TTL
// Redis: SET auth:refresh:{jti} {userID} EX {ttlSeconds}
func (r *TokenCacheRepository) StoreRefreshToken(ctx context.Context, jti string, userID int64, ttl time.Duration) error {
	cmd := r.CacheClient.B().Set().
		Key(fmt.Sprintf(refreshTokenPrefix, jti)).
		Value(strconv.FormatInt(userID, 10)).
		Ex(ttl).
		Build()

	if err := r.CacheClient.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("store refresh token [%s]: %w", jti, err)
	}
	return nil
}

// CheckRefreshToken verifies whether the refreshToken is still valid (residing in Redis)
// Redis: GET auth:refresh:{jti} → userID
func (r *TokenCacheRepository) CheckRefreshToken(ctx context.Context, jti string) (int64, error) {
	cmd := r.CacheClient.B().Get().
		Key(fmt.Sprintf(refreshTokenPrefix, jti)).
		Build()

	val, err := r.CacheClient.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return 0, errors.New("refresh token not found or expired")
		}
		return 0, fmt.Errorf("check refresh token [%s]: %w", jti, err)
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse userID from refresh token [%s]: %w", jti, err)
	}
	return userID, nil
}

// RevokeRefreshToken Revoke refreshToken (used for Refresh Rotation or active cancellation)
// Redis: DEL auth:refresh:{jti}
func (r *TokenCacheRepository) RevokeRefreshToken(ctx context.Context, jti string) error {
	cmd := r.CacheClient.B().Del().
		Key(fmt.Sprintf(refreshTokenPrefix, jti)).
		Build()

	if err := r.CacheClient.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("revoke refresh token [%s]: %w", jti, err)
	}
	return nil
}

// BlacklistAccessToken adds the JTI of accessToken to the blacklist, TTL = token remaining validity period
// Redis: SET auth:blacklist:{jti} "1" EX {ttlSeconds}
func (r *TokenCacheRepository) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	cmd := r.CacheClient.B().Set().
		Key(fmt.Sprintf(accessTokenPrefix, jti)).
		Value("1").
		Ex(ttl).
		Build()

	if err := r.CacheClient.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("blacklist access token [%s]: %w", jti, err)
	}
	return nil
}

// IsBlacklisted Check if the accessToken's JTI is on the blacklist.
// Redis: EXISTS auth:blacklist:{jti}
func (r *TokenCacheRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	cmd := r.CacheClient.B().Exists().
		Key(fmt.Sprintf(accessTokenPrefix, jti)).
		Build()

	n, err := r.CacheClient.Do(ctx, cmd).AsInt64()
	if err != nil {
		return false, fmt.Errorf("check blacklist for token [%s]: %w", jti, err)
	}
	return n > 0, nil
}

var _ TokenRepository = (*TokenCacheRepository)(nil)
