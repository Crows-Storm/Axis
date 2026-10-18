package service

import (
	"context"

	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/common/server/metrics"
	"github.com/Crows-Storm/Axis/exchange/adapter"
	"github.com/Crows-Storm/Axis/exchange/adapter/binance"
	"github.com/Crows-Storm/Axis/exchange/adapter/bitget"
	"github.com/Crows-Storm/Axis/exchange/adapter/okx"
	"github.com/Crows-Storm/Axis/exchange/app"
	"github.com/Crows-Storm/Axis/exchange/app/command"
	"github.com/Crows-Storm/Axis/exchange/app/query"
)

type ApplicationDependencies struct {
	CacheClient cache.RueidisClient
	Issuer      jwt.TokenIssuer
}

func NewApplication(_ context.Context, _ ApplicationDependencies) (app.Application, func()) {
	return newApplication(), func() {}
}

func newApplication() app.Application {
	venues := adapter.NewRegistry(binance.NewClient(), okx.NewClient(), bitget.NewClient())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			PlaceOrder:  command.NewPlaceOrderHandler(venues, metricsClient),
			CancelOrder: command.NewCancelOrderHandler(venues, metricsClient),
		},
		Queries: app.Queries{
			GetAccountInfo: query.NewGetAccountInfoHandler(venues, metricsClient),
			GetBalance:     query.NewGetBalanceHandler(venues, metricsClient),
			GetTicker:      query.NewGetTickerHandler(venues, metricsClient),
			GetCandles:     query.NewGetCandlesHandler(venues, metricsClient),
			GetOrderBook:   query.NewGetOrderBookHandler(venues, metricsClient),
			GetOrder:       query.NewGetOrderHandler(venues, metricsClient),
		},
	}
}
