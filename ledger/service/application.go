package service

import (
	"context"

	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/ledger/app"
)

func NewApplication(
	ctx context.Context,
	cacheClient cache.RueidisClient,
) (app.Application, func()) {

	return newApplication(ctx, cacheClient), func() {
		// nothing
	}
}

func newApplication(
	_ context.Context,
	_ cache.RueidisClient,
) app.Application {

	//metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
