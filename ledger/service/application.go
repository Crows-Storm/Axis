package service

import (
	"context"

	redisPkg "github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/ledger/app"
	"github.com/sirupsen/logrus"
)

func NewApplication(
	ctx context.Context,
	logger *logrus.Entry,
	cacheClient redisPkg.RueidisClient,
) (app.Application, func()) {

	return newApplication(ctx, logger, cacheClient), func() {
		// nothing
	}
}

func newApplication(
	_ context.Context,
	_ *logrus.Entry,
	_ redisPkg.RueidisClient,
) app.Application {

	//metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
