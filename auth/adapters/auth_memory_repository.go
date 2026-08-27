package adapters

import (
	"context"
	"sync"

	"github.com/Crows-Storm/Axis/common/domain/principal"
)

type MemoryAuthRepository struct {
	lock  *sync.RWMutex
	store []*principal.Principal
}

func NewMemoryAuthRepository() *MemoryAuthRepository {
	return &MemoryAuthRepository{
		lock:  &sync.RWMutex{},
		store: make([]*principal.Principal, 0),
	}
}

func (a MemoryAuthRepository) GetPrincipal(ctx context.Context, token string) (*principal.Principal, error) {
	//TODO implement me
	panic("implement me")
}
