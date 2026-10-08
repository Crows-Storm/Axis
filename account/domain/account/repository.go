package account

import (
	"context"

	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
)

type ListFilter struct {
	UserID uint64
	Type   commwallet.Type
}

type Repository interface {
	Get(ctx context.Context, id uint64) (*Account, error)
	GetByWalletID(ctx context.Context, walletID uint64) (*Account, error)
	List(ctx context.Context, filter ListFilter) ([]*Account, error)
	Save(ctx context.Context, acc *Account) error
	ListCurrencies(ctx context.Context) ([]commaccount.Currency, error)
	GetCurrency(ctx context.Context, id uint64, symbol string) (commaccount.Currency, error)
}
