package service

import (
	"context"

	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/wallet/app"
)

func NewApplication(
	ctx context.Context,
	_ cache.RueidisClient,
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
