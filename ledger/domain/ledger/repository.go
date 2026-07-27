package ledger

import "context"

type Repository interface {
	GetLedgerByWalletId(ctx context.Context, walletId int64) (*Ledger, error)
}
