package service

import (
	"context"

	"github.com/Crows-Storm/Axis/wallet/app"
)

func NewApplication(ctx context.Context) app.Application {
	//walletRepo := adapters.NewMemoryWalletRepository()
	//logger := logrus.NewEntry(logrus.New())
	//metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
