package query

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
)

type GetAccountQuery struct {
	ID uint64
}

type GetAccountQueryHandler decorator.QueryHandler[GetAccountQuery, *domain.Account]

func NewGetAccountQueryHandler(repo domain.Repository, metrics decorator.MetricsClient) GetAccountQueryHandler {
	if repo == nil {
		panic("nil account repository")
	}
	return decorator.ApplyQueryDecorators[GetAccountQuery, *domain.Account](
		getAccountQueryHandler{repo: repo},
		metrics,
	)
}

type getAccountQueryHandler struct {
	repo domain.Repository
}

func (h getAccountQueryHandler) Handle(ctx context.Context, q GetAccountQuery) (*domain.Account, error) {
	if q.ID == 0 {
		return nil, fmt.Errorf("get account: %w", commaccount.ErrInvalid)
	}
	acc, err := h.repo.Get(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	return acc, nil
}
