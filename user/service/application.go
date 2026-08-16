package service

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/metrics"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/common/server/store"
	"github.com/Crows-Storm/Axis/user/adapters"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

// ApplicationDependencies Includes all external dependencies required by the application.
type ApplicationDependencies struct {
	Store       *store.Store
	CacheClient cache.RueidisClient
}

// NewApplication Create a new Application Instance
func NewApplication(
	ctx context.Context,
	deps ApplicationDependencies,
) (app.Application, func()) {
	userRepo := createUserRepository(deps.Store)
	metricsClient := metrics.TodoMetrics{}

	application := newApplication(ctx, userRepo, deps.CacheClient, metricsClient)

	cleanup := func() {
		// some close func
	}

	return application, cleanup
}

// createUserRepository Create a suitable repository implementation based on the configuration
func createUserRepository(st *store.Store) domain.Repository {
	return adapters.NewUserMariaRepository(st)
}

// newApplication Builder Application layer（Commands + Queries）
func newApplication(
	_ context.Context,
	userRepo domain.Repository,
	_ cache.RueidisClient,
	metricsClient decorator.MetricsClient,
) app.Application {
	return app.Application{
		Commands: app.Commands{
			CreateUser:     command.NewCreateUserCommandHandler(userRepo, metricsClient),
			UpdateUser:     command.NewUpdateUserCommandHandler(userRepo, metricsClient),
			SoftDeleteUser: command.NewSoftDeleteUserCommandHandler(userRepo, metricsClient),
			DisableUser:    command.NewDisableUserCommandHandler(userRepo, metricsClient),
		},
		Queries: app.Queries{
			GetUser:            query.NewGetUserQueryHandler(userRepo, metricsClient),
			UserExists:         query.NewUserExistsHandler(userRepo, metricsClient),
			UserStatusAnalysis: query.NewUserStatusAnalysisHandler(userRepo, metricsClient),
		},
	}
}
