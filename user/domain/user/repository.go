package user

import (
	"context"
	"fmt"
)

type Repository interface {
	GetInfo(id int64) (*User, error)
	GetByLoginId(ctx context.Context, loginId string) (*User, error)
	ExistsWithTransaction(ctx context.Context, id int64, loginId string, email string) (bool, error)
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetPasswordByLoginId(ctx context.Context, loginId string) string

	Create(ctx context.Context, user *User) (*User, error)
	CreateBatch(ctx context.Context, users []*User) error
	Update(
		ctx context.Context,
		user *User,
		updateFun func(context.Context, *User) (*User, error),
	) error
	Disable(ctx context.Context, userId int64) error

	// Dangerous operation
	SoftDelete(ctx context.Context, userId int64) error
}

type NotFoundError struct {
	UserId int64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("User %d Not Found !!!", e.UserId)
}
