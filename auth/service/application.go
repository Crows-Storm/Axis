package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/adapters/grpc"
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/auth/app/query"
	"github.com/Crows-Storm/Axis/auth/app/service/auth"
	"github.com/Crows-Storm/Axis/common/client"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/common/server/store"
)

type ApplicationDependencies struct {
	Store         *store.Store
	CacheClient   cache.RueidisClient
	Issuer        *jwt.JWTIssuer
	MetricsClient decorator.MetricsClient
}

func NewApplication(
	ctx context.Context,
	deps ApplicationDependencies,
) (app.Application, func()) {

	userGrpcClient, closeUserGRPCClient, err := client.NewUserGRPCClient(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create user gRPC client")
	}

	// TODO: need init GRPC service
	var userGRPC = grpc.NewUserGRPC(userGrpcClient)

	// The authService is Authorization Implementation Strategy Context Structure
	authService := auth.InitAuthAppService(userGRPC, deps.Issuer)

	return newApplication(ctx, userGRPC, deps.CacheClient, authService, deps.MetricsClient), func() {
		_ = closeUserGRPCClient()
	}
}

func newApplication(
	_ context.Context,
	userService provider.UserService,
	cacheClient cache.RueidisClient,
	authService *auth.AuthAppService,
	metricsClient decorator.MetricsClient,
) app.Application {

	return app.Application{
		Commands: app.Commands{
			RegisterUser: command.NewRegisterUserCommandHandler(
				userService,
				cacheClient,
				metricsClient,
			),
			Login: command.NewLoginCommandHandler(userService, authService, metricsClient),
			//Logout: command.RegisterAndLoginCommandHandler,
			//RefreshToken     command.NewRefreshTokenCommandHandler()
			Logout: command.NewLogoutCommandHandler(authService, metricsClient),

			//PasswordLoginCommand: command.NewRegisterUserCommandHandler(
			//	AuthService: authService,
			//)
		},
		Queries: app.Queries{
			GetUser: query.NewGetUserQueryHandler(userService, metricsClient),
		},
	}
}
