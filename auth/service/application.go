package service

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/adapters"
	"github.com/Crows-Storm/Axis/auth/adapters/grpc"
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/auth/app/provider"
	"github.com/Crows-Storm/Axis/auth/app/query"
	"github.com/Crows-Storm/Axis/auth/app/service/auth"
	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/auth/domain/role"
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
	EventBus      decorator.EventBus
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

	roleRepo := adapters.NewRoleMariaRepository(deps.Store)
	menuRepo := adapters.NewMenuMariaRepository(deps.Store)
	permRepo := adapters.NewPermissionMariaRepository(deps.Store)

	return newApplication(ctx, userGRPC, deps.CacheClient, authService, roleRepo, menuRepo, permRepo, deps.MetricsClient, deps.EventBus), func() {
		_ = closeUserGRPCClient()
		if deps.EventBus != nil {
			_ = deps.EventBus.Close()
		}
	}
}

func newApplication(
	_ context.Context,
	userService provider.UserService,
	cacheClient cache.RueidisClient,
	authService *auth.AuthAppService,
	roleRepo role.Repository,
	menuRepo menu.Repository,
	permRepo permission.Repository,
	metricsClient decorator.MetricsClient,
	bus decorator.EventBus,
) app.Application {

	return app.Application{
		Commands: app.Commands{
			RegisterUser: command.NewRegisterUserCommandHandler(
				userService,
				cacheClient,
				metricsClient,
			),
			Login:              command.NewLoginCommandHandler(userService, authService, roleRepo, metricsClient),
			Logout:             command.NewLogoutCommandHandler(authService, metricsClient),
			AssignUserRole:     command.NewAssignUserRoleCommandHandler(roleRepo, metricsClient, bus),
			UnassignUserRole:   command.NewUnassignUserRoleCommandHandler(roleRepo, metricsClient, bus),
			ReplaceUserRoles:   command.NewReplaceUserRolesCommandHandler(roleRepo, metricsClient, bus),
			UpdateRoleMenuMask: command.NewUpdateRoleMenuMaskCommandHandler(roleRepo, metricsClient, bus),
			RevokeRoleMenu:     command.NewRevokeRoleMenuCommandHandler(roleRepo, metricsClient, bus),
			CreateRole:         command.NewCreateRoleCommandHandler(roleRepo, metricsClient, bus),
			UpdateRole:         command.NewUpdateRoleCommandHandler(roleRepo, metricsClient, bus),
			UpdateRoleStatus:   command.NewUpdateRoleStatusCommandHandler(roleRepo, metricsClient, bus),
			CreateMenu:         command.NewCreateMenuCommandHandler(menuRepo, metricsClient, bus),
			UpdateMenu:         command.NewUpdateMenuCommandHandler(menuRepo, metricsClient, bus),
			UpdateMenuStatus:   command.NewUpdateMenuStatusCommandHandler(menuRepo, metricsClient, bus),
		},
		Queries: app.Queries{
			GetUser:             query.NewGetUserQueryHandler(userService, metricsClient),
			GetUserRoles:        query.NewGetUserRolesQueryHandler(roleRepo, metricsClient),
			GetUserMenuTree:     query.NewGetUserMenuTreeQueryHandler(menuRepo, permRepo, metricsClient),
			CheckMenuPermission: query.NewCheckMenuPermissionQueryHandler(permRepo, metricsClient),
			ListRoles:           query.NewListRolesQueryHandler(roleRepo, metricsClient),
			GetRole:             query.NewGetRoleQueryHandler(roleRepo, menuRepo, metricsClient),
			ListAssignedRoles:   query.NewListAssignedRolesQueryHandler(roleRepo, metricsClient),
			ListMenuCatalog:     query.NewListMenuCatalogQueryHandler(menuRepo, metricsClient),
			ListActions:         query.NewListActionsQueryHandler(permRepo, metricsClient),
			GetMenu:             query.NewGetMenuQueryHandler(menuRepo, metricsClient),
		},
	}
}
