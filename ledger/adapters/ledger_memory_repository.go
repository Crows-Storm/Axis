package adapters

import (
	"context"
	"sync"

	domain "github.com/Crows-Storm/Axis/ledger/domain/ledger"
)

type LedgerMemoryRepository struct {
	lock  *sync.RWMutex
	store []*domain.Ledger
}

func NewLedgerMemoryRepository() *LedgerMemoryRepository {
	return &LedgerMemoryRepository{
		lock:  &sync.RWMutex{},
		store: make([]*domain.Ledger, 0),
	}
}

func (l LedgerMemoryRepository) GetLedgerByWalletId(ctx context.Context, walletId int64) (*domain.Ledger, error) {
	//TODO implement me
	panic("implement me")
}
