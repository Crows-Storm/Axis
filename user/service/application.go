package service

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/metrics"
	redisPkg "github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/user/adapters"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
)

func NewApplication(
	ctx context.Context,
	logger *logrus.Entry,
	cacheClient redisPkg.RueidisClient,
) (app.Application, func()) {
	userRepo := adapters.NewMemoryUserRepository()
	metricsClient := metrics.TodoMetrics{}

	return newApplication(ctx, userRepo, cacheClient, logger, metricsClient), func() {
		// TODO: nothing
	}
}

func newApplication(
	_ context.Context,
	userRepo domain.Repository,
	cacheClient redisPkg.RueidisClient,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) app.Application {
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
