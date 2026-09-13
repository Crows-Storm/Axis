package user

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	QueryRepository

	CreatePreCheckRepository
	BatchRegistrationRepository

	RegistrationRepository

	ProfileRepository

	DeleteRepository
}

type ProfileRepository interface {
	Update(ctx context.Context, user *User) error
	Disable(ctx context.Context, userId int64) error
}

type DeleteRepository interface {
	// Dangerous operation
	SoftDelete(ctx context.Context, userId int64) error
}

type RegistrationRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
}

type BatchRegistrationRepository interface {
	CreateBatch(ctx context.Context, users []*User) error
}

type QueryRepository interface {
	GetInfo(id int64) (*User, error)
	GetByLoginId(ctx context.Context, loginId string) (*User, error)
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetPasswordByLoginId(ctx context.Context, loginId string) string
}

type CreatePreCheckRepository interface {
	ExistsWithTransaction(ctx context.Context, id int64, loginId string, email string) (bool, error)
}

type NotFoundError struct {
	UserId int64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("User %d Not Found !!!", e.UserId)
}

func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 0")
}
