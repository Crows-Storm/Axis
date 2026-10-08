package query

import (
	"context"
	"fmt"

	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
	commaccount "github.com/Crows-Storm/Axis/common/domain/account"
)

type ListCurrenciesQuery struct{}

type ListCurrenciesQueryHandler decorator.QueryHandler[ListCurrenciesQuery, []commaccount.Currency]

func NewListCurrenciesQueryHandler(repo domain.Repository, metrics decorator.MetricsClient) ListCurrenciesQueryHandler {
	if repo == nil {
		panic("nil account repository")
	}
	return decorator.ApplyQueryDecorators[ListCurrenciesQuery, []commaccount.Currency](
		listCurrenciesQueryHandler{repo: repo},
		metrics,
	)
}

type listCurrenciesQueryHandler struct {
	repo domain.Repository
}

func (h listCurrenciesQueryHandler) Handle(ctx context.Context, _ ListCurrenciesQuery) ([]commaccount.Currency, error) {
	rows, err := h.repo.ListCurrencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list currencies: %w", err)
	}
	return rows, nil
}
