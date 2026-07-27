package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/adapters/grpc"
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/common/client"
	"github.com/Crows-Storm/Axis/common/metrics"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	grpcClient, closeUserGRPCClient, err := client.NewUserGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	userGRPC := grpc.NewUserGRPC(grpcClient)
	return newApplication(ctx, userGRPC), func() {
		_ = closeUserGRPCClient()
	}
}

func newApplication(_ context.Context, userGRPC command.UserService) app.Application {
	//authRepo := adapters.NewMemoryAuthRepository()
	logger := logrus.NewEntry(logrus.New())
	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			RegisterUser: command.NewRegisterUserCommandHandler(userGRPC, logger, metricsClient),
		},
		Queries: app.Queries{},
	}
}
