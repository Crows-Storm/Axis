package user

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/pagination"
	"gorm.io/gorm"
)

type PageFilter struct {
	LoginID string
	Email   string
	Status  *int8
}

type Repository interface {
	GetInfo(id uint64) (*User, error)
	GetByLoginID(ctx context.Context, loginID string) (*User, error)
	ExistsWithTransaction(ctx context.Context, id uint64, loginID string, email string) (bool, error)
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetPasswordByLoginID(ctx context.Context, loginID string) string
	Page(ctx context.Context, filter PageFilter, args pagination.Args) (*pagination.Connection[*User], error)

	Create(ctx context.Context, user *User) (*User, error)
	CreateBatch(ctx context.Context, users []*User) error
	Update(
		ctx context.Context,
		user *User,
		updateFun func(context.Context, *User) (*User, error),
	) error
	StatusChange(context.Context, *User) error

	// Dangerous operation
	SoftDelete(ctx context.Context, id uint64) error

	// User Profile
	SaveProfile(ctx context.Context, user *User) error
}

type NotFoundError struct {
	UserId uint64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("User %d Not Found !!!", e.UserId)
}

func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 0")
}
