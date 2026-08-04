package service

import (
	"context"

	redisPkg "github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/ledger/app"
)

func NewApplication(
	ctx context.Context,
	cacheClient redisPkg.RueidisClient,
) (app.Application, func()) {

	return newApplication(ctx, cacheClient), func() {
		// nothing
	}
}

func newApplication(
	_ context.Context,
	_ redisPkg.RueidisClient,
) app.Application {

	//metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
