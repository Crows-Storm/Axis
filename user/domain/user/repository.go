package user

import (
	"context"
	"fmt"
)

type Repository interface {
	GetInfo(id uint64) (*User, error)
	GetByLoginID(ctx context.Context, loginID string) (*User, error)
	ExistsWithTransaction(ctx context.Context, id uint64, loginID string, email string) (bool, error)
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetPasswordByLoginID(ctx context.Context, loginID string) string

	Create(ctx context.Context, user *User) (*User, error)
	CreateBatch(ctx context.Context, users []*User) error
	Update(
		ctx context.Context,
		user *User,
		updateFun func(context.Context, *User) (*User, error),
	) error
	Disable(ctx context.Context, id uint64) error

	// Dangerous operation
	SoftDelete(ctx context.Context, id uint64) error
}

type NotFoundError struct {
	UserId uint64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("User %d Not Found !!!", e.UserId)
}
