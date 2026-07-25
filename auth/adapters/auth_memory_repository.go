package adapters

import (
	"context"
	"sync"

	"github.com/Crows-Storm/Axis/auth/domain"
)

type MemoryAuthRepository struct {
	lock  *sync.RWMutex
	store []*domain.Principal
}

func NewMemoryAuthRepository() *MemoryAuthRepository {
	return &MemoryAuthRepository{
		lock:  &sync.RWMutex{},
		store: make([]*domain.Principal, 0),
	}
}

func (a MemoryAuthRepository) GetPrincipal(ctx context.Context, token string) (*domain.Principal, error) {
	//TODO implement me
	panic("implement me")
}
