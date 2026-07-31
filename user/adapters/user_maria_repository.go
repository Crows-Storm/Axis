package adapters

import (
	"context"
	"sync"

	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type UserMariaRepository struct {
	lock *sync.RWMutex
}

// NewUserMariaRepository TODO: need add a Transaction of DB? maybe lock from outside import? this is a good question
func NewUserMariaRepository() *UserMariaRepository {
	return &UserMariaRepository{lock: &sync.RWMutex{}}
}

func (u UserMariaRepository) GetInfo(id int64) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserMariaRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserMariaRepository) Update(ctx context.Context, user *domain.User, updateFun func(context.Context, *domain.User) (*domain.User, error)) error {
	//TODO implement me
	panic("implement me")
}
