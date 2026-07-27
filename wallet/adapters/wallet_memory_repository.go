package adapters

import (
	"context"
	"sync"

	domain "github.com/Crows-Storm/Axis/wallet/domain/wallet"
)

type MemoryWalletRepository struct {
	lock  *sync.RWMutex
	store []*domain.Wallet
}

func NewMemoryWalletRepository() *MemoryWalletRepository {
	return &MemoryWalletRepository{
		lock:  &sync.RWMutex{},
		store: make([]*domain.Wallet, 0),
	}
}

func (m MemoryWalletRepository) GetWallet(ctx context.Context, id int64) (*domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}

func (m MemoryWalletRepository) GetWallets(ctx context.Context, userId int64) ([]*domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}
