package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/adapters/grpc"
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/common/client"
	"github.com/Crows-Storm/Axis/common/metrics"
	redisPkg "github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/sirupsen/logrus"
)

func NewApplication(
	ctx context.Context,
	logger *logrus.Entry,
	cacheClient redisPkg.RueidisClient,
) (app.Application, func()) {

	grpcClient, closeUserGRPCClient, err := client.NewUserGRPCClient(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create user gRPC client")
	}

	userGRPC := grpc.NewUserGRPC(grpcClient)

	return newApplication(ctx, userGRPC, cacheClient, logger), func() {
		_ = closeUserGRPCClient()
	}
}

func newApplication(
	_ context.Context,
	userGRPC command.UserService,
	cacheClient redisPkg.RueidisClient,
	logger *logrus.Entry,
) app.Application {

	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			RegisterUser: command.NewRegisterUserCommandHandler(
				userGRPC,
				cacheClient,
				logger,
				metricsClient,
			),
		},
		Queries: app.Queries{},
	}
}
