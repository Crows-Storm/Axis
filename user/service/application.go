package service

import (
	"context"

	"github.com/Crows-Storm/Axis/common/metrics"
	"github.com/Crows-Storm/Axis/user/adapters"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	userRepo := adapters.NewMemoryUserRepository()
	logger := logrus.NewEntry(logrus.New())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateUser: command.NewCreateUserCommandHandler(userRepo, logger, metricsClient),
			UpdateUser: command.NewUpdateUserCommandHandler(userRepo, logger, metricsClient),
		},
		Queries: app.Queries{
			GetUser: query.NewGetUserQueryHandler(userRepo, logger, metricsClient),
		},
	}
}
