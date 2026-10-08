package service

import (
	"context"
	"log"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/exchange/adapter/binance"
	"github.com/Crows-Storm/Axis/exchange/app"
)

type ApplicationDependencies struct {
	CacheClient cache.RueidisClient
	Issuer      jwt.TokenIssuer
	EventBus    decorator.EventBus
}

func NewApplication(ctx context.Context, dependency ApplicationDependencies) (app.Application, func()) {
	return newApplication(ctx), func() {

	}
}

func newApplication(_ context.Context) app.Application {
	binanceClient := binance.NewClient()
	// TODO: init all exchange client
	log.Fatal(binanceClient)
	return app.Application{}
}
