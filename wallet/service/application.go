package service

import (
	"context"

	redisPkg "github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/wallet/app"
)

func NewApplication(
	ctx context.Context,
	_ redisPkg.RueidisClient,
) (app.Application, func()) {
	//walletRepo := adapters.NewMemoryWalletRepository()
	//logger := logrus.NewEntry(logrus.New())
	//metricsClient := metrics.TodoMetrics{}
	return newApplication(ctx), func() {
		// nothing
	}
}

func newApplication(ctx context.Context) app.Application {

	//metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
