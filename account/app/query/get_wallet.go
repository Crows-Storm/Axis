package query

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
)

type GetWalletQuery struct {
	ID uint64
}

type WalletView struct {
	Account *domain.Account
	Wallet  domain.Wallet
}

type GetWalletQueryHandler decorator.QueryHandler[GetWalletQuery, WalletView]

func NewGetWalletQueryHandler(repo domain.Repository, metrics decorator.MetricsClient) GetWalletQueryHandler {
	if repo == nil {
		panic("nil account repository")
	}
	return decorator.ApplyQueryDecorators[GetWalletQuery, WalletView](
		getWalletQueryHandler{repo: repo},
		metrics,
	)
}

type getWalletQueryHandler struct {
	repo domain.Repository
}

func (h getWalletQueryHandler) Handle(ctx context.Context, q GetWalletQuery) (WalletView, error) {
	if q.ID == 0 {
		return WalletView{}, fmt.Errorf("get wallet: %w", commaccount.ErrInvalid)
	}
	acc, err := h.repo.GetByWalletID(ctx, q.ID)
	if err != nil {
		return WalletView{}, fmt.Errorf("get wallet: %w", err)
	}
	wallet, ok := acc.WalletByID(q.ID)
	if !ok {
		return WalletView{}, fmt.Errorf("get wallet: %w", commaccount.ErrNotFound)
	}
	return WalletView{Account: acc, Wallet: wallet}, nil
}
