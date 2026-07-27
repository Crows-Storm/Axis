package wallet

import (
	"context"
)

type Repository interface {
	GetWallet(ctx context.Context, id int64) (*Wallet, error)
	GetWallets(ctx context.Context, userId int64) ([]*Wallet, error)
}
