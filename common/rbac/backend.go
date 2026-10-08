package rbac

import "context"

// Backend is the Redis (or memory) surface used by the authorizer.
type Backend interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string) error
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SAdd(ctx context.Context, key string, members ...string) error
	HGet(ctx context.Context, key, field string) (string, bool, error)
	HSet(ctx context.Context, key string, fields map[string]string) error
}

// Source loads RBAC rows from MariaDB. Wallet and other services leave it nil and only read Redis.
type Source interface {
	EnabledRoleIDs(ctx context.Context, userID uint64) ([]uint64, error)
	RoleMenuMasks(ctx context.Context, roleID uint64) (map[string]uint64, error)
	RoleIDsForMenuCode(ctx context.Context, menuCode string) ([]uint64, error)
}
