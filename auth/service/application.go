package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/app"
)

func NewApplication(ctx context.Context) app.Application {
	//userRepo := adapters.NewMemoryAuthRepository()
	//logger := logrus.NewEntry(logrus.New())
	//metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
