package user

import (
	"context"
	"fmt"
)

type Repository interface {
	GetInfo(id int64) (*User, error)

	Create(ctx context.Context, user *User) (*User, error)
	Update(
		ctx context.Context,
		user *User,
		updateFun func(context.Context, *User) (*User, error),
	) error
}

type NotFoundError struct {
	UserId int64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("User %d Not Found !!!", e.UserId)
}
