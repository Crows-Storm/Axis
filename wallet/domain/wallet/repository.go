package wallet

import (
	"context"
)

type Repository interface {
	GetWallet(ctx context.Context, id uint64) (*Wallet, error)
	GetWallets(ctx context.Context, userId uint64) ([]*Wallet, error)
}
