package service

import (
	"context"

	"github.com/Crows-Storm/Axis/account/adapters"
	"github.com/Crows-Storm/Axis/account/app"
	"github.com/Crows-Storm/Axis/account/app/command"
	"github.com/Crows-Storm/Axis/account/app/query"
	domain "github.com/Crows-Storm/Axis/account/domain/account"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/server/metrics"
	"github.com/Crows-Storm/Axis/common/server/store"
)

type ApplicationDependencies struct {
	Store    *store.Store
	EventBus decorator.EventBus
}

func NewApplication(_ context.Context, deps ApplicationDependencies) (app.Application, func()) {
	repo := createRepository(deps.Store)
	metricsClient := metrics.TodoMetrics{}
	application := app.Application{
		Commands: app.Commands{
			OpenAccount: command.NewOpenAccountCommandHandler(repo, metricsClient, deps.EventBus),
			OpenWallet:  command.NewOpenWalletCommandHandler(repo, metricsClient, deps.EventBus),
		},
		Queries: app.Queries{
			GetAccount:     query.NewGetAccountQueryHandler(repo, metricsClient),
			ListAccounts:   query.NewListAccountsQueryHandler(repo, metricsClient),
			GetWallet:      query.NewGetWalletQueryHandler(repo, metricsClient),
			ListCurrencies: query.NewListCurrenciesQueryHandler(repo, metricsClient),
		},
	}
	cleanup := func() {
		if deps.EventBus != nil {
			_ = deps.EventBus.Close()
		}
	}
	return application, cleanup
}

func createRepository(st *store.Store) domain.Repository {
	return adapters.NewAccountMariaRepository(st)
}
