package ledger

import "context"

type Repository interface {
	GetLedgerByWalletId(ctx context.Context, walletId uint64) (*Ledger, error)
}
