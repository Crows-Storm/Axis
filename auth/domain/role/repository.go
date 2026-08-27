package role

import (
	"context"
)

type Repository interface {
	// Get Role by id
	Get(ctx context.Context, id int64) (*Role, error)
	GetByUserId(ctx context.Context, userId int64) (*Role, error)
}
