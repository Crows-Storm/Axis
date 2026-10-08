package query

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
	commwallet "github.com/Crows-Storm/Axis/common/domain/wallet"
)

type ListAccountsQuery struct {
	UserID uint64
	Type   string
}

type ListAccountsQueryHandler decorator.QueryHandler[ListAccountsQuery, []*domain.Account]

func NewListAccountsQueryHandler(repo domain.Repository, metrics decorator.MetricsClient) ListAccountsQueryHandler {
	if repo == nil {
		panic("nil account repository")
	}
	return decorator.ApplyQueryDecorators[ListAccountsQuery, []*domain.Account](
		listAccountsQueryHandler{repo: repo},
		metrics,
	)
}

type listAccountsQueryHandler struct {
	repo domain.Repository
}

func (h listAccountsQueryHandler) Handle(ctx context.Context, q ListAccountsQuery) ([]*domain.Account, error) {
	if q.UserID == 0 {
		return nil, fmt.Errorf("list accounts: %w", commaccount.ErrInvalid)
	}
	filter := domain.ListFilter{UserID: q.UserID}
	if q.Type != "" {
		kind, ok := commwallet.ParseType(q.Type)
		if !ok {
			return nil, fmt.Errorf("list accounts: %w", commaccount.ErrInvalid)
		}
		filter.Type = kind
	}
	accounts, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return accounts, nil
}
