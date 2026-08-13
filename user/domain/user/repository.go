package user

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	GetInfo(id int64) (*User, error)

	Create(ctx context.Context, user *User) (*User, error)
	CreateBatch(ctx context.Context, users []*User) error
	Update(
		ctx context.Context,
		user *User,
		updateFun func(context.Context, *User) (*User, error),
	) error
	UpdateStatus(ctx context.Context, userId int64, status int8) error

	SoftDelete(ctx context.Context, userId int64) error
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*User, int64, error)
	ExistsWithTransaction(tx *gorm.DB, userId int64) (bool, error)
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

type NotFoundError struct {
	UserId int64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("User %d Not Found !!!", e.UserId)
}
