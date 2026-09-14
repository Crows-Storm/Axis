package role

import (
	"context"
)

type Repository interface {
	// Get Role by id
	Get(ctx context.Context, id uint64) (*Role, error)
	GetByUserId(ctx context.Context, userId uint64) (*Role, error)
}
