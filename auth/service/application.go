package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/adapters/grpc"
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/common/client"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/metrics"
	"github.com/Crows-Storm/Axis/common/server/redis"
)

func NewApplication(
	ctx context.Context,
	cacheClient redis.RueidisClient,
) (app.Application, func()) {

	grpcClient, closeUserGRPCClient, err := client.NewUserGRPCClient(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create user gRPC client")
	}

	userGRPC := grpc.NewUserGRPC(grpcClient)

	return newApplication(ctx, userGRPC, cacheClient), func() {
		_ = closeUserGRPCClient()
	}
}

func newApplication(
	_ context.Context,
	userGRPC command.UserService,
	cacheClient redis.RueidisClient,
) app.Application {

	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			RegisterUser: command.NewRegisterUserCommandHandler(
				userGRPC,
				cacheClient,
				metricsClient,
			),
		},
		Queries: app.Queries{},
	}
}
