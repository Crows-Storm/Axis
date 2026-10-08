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
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/common/server/store"
)

type ApplicationDependencies struct {
	Store         *store.Store
	CacheClient   cache.RueidisClient
	Issuer        *jwt.JWTIssuer
	MetricsClient decorator.MetricsClient
	EventBus      decorator.EventBus
	Authorizer    *rbac.Authorizer
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
	authz := deps.Authorizer
	if authz == nil {
		authz = rbac.NewAuthorizer(rbac.NewRedis(deps.CacheClient), adapters.NewMariaRBACSource(deps.Store))
	}

	return newApplication(ctx, userGRPC, deps.CacheClient, authService, roleRepo, menuRepo, permRepo, authz, deps.MetricsClient, deps.EventBus), func() {
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
	authz *rbac.Authorizer,
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
			Login: loginAndWarm{
				inner: command.NewLoginCommandHandler(userService, authService, roleRepo, metricsClient),
				authz: authz,
			},
			Logout: command.NewLogoutCommandHandler(authService, metricsClient),
			AssignUserRole: afterOK(command.NewAssignUserRoleCommandHandler(roleRepo, metricsClient, bus), func(ctx context.Context, cmd command.AssignUserRoleCommand) {
				_ = authz.InvalidateUser(ctx, cmd.UserID)
			}),
			UnassignUserRole: afterOK(command.NewUnassignUserRoleCommandHandler(roleRepo, metricsClient, bus), func(ctx context.Context, cmd command.UnassignUserRoleCommand) {
				_ = authz.InvalidateUser(ctx, cmd.UserID)
			}),
			ReplaceUserRoles: afterOK(command.NewReplaceUserRolesCommandHandler(roleRepo, metricsClient, bus), func(ctx context.Context, cmd command.ReplaceUserRolesCommand) {
				_ = authz.InvalidateUser(ctx, cmd.UserID)
			}),
			UpdateRoleMenuMask: afterOK(command.NewUpdateRoleMenuMaskCommandHandler(roleRepo, metricsClient, bus), func(ctx context.Context, cmd command.UpdateRoleMenuMaskCommand) {
				_ = authz.InvalidateRole(ctx, cmd.RoleID)
			}),
			RevokeRoleMenu: afterOK(command.NewRevokeRoleMenuCommandHandler(roleRepo, metricsClient, bus), func(ctx context.Context, cmd command.RevokeRoleMenuCommand) {
				_ = authz.InvalidateRole(ctx, cmd.RoleID)
			}),
			CreateRole: command.NewCreateRoleCommandHandler(roleRepo, metricsClient, bus),
			UpdateRole: command.NewUpdateRoleCommandHandler(roleRepo, metricsClient, bus),
			UpdateRoleStatus: afterOK(command.NewUpdateRoleStatusCommandHandler(roleRepo, metricsClient, bus), func(ctx context.Context, cmd command.UpdateRoleStatusCommand) {
				_ = authz.InvalidateRole(ctx, cmd.RoleID)
			}),
			CreateMenu: command.NewCreateMenuCommandHandler(menuRepo, metricsClient, bus),
			UpdateMenu: afterOK(command.NewUpdateMenuCommandHandler(menuRepo, metricsClient, bus), func(ctx context.Context, cmd command.UpdateMenuCommand) {
				current, err := menuRepo.Get(ctx, cmd.MenuID)
				if err == nil {
					_ = authz.InvalidateMenu(ctx, current.Code)
				}
			}),
			UpdateMenuStatus: afterOK(command.NewUpdateMenuStatusCommandHandler(menuRepo, metricsClient, bus), func(ctx context.Context, cmd command.UpdateMenuStatusCommand) {
				current, err := menuRepo.Get(ctx, cmd.MenuID)
				if err == nil {
					_ = authz.InvalidateMenu(ctx, current.Code)
				}
			}),
		},
		Queries: app.Queries{
			GetUser:             query.NewGetUserQueryHandler(userService, metricsClient),
			GetUserRoles:        query.NewGetUserRolesQueryHandler(roleRepo, metricsClient),
			GetUserMenuTree:     query.NewGetUserMenuTreeQueryHandler(menuRepo, permRepo, metricsClient),
			CheckMenuPermission: query.NewCheckMenuPermissionQueryHandler(authz, metricsClient),
			ListRoles:           query.NewListRolesQueryHandler(roleRepo, metricsClient),
			GetRole:             query.NewGetRoleQueryHandler(roleRepo, menuRepo, metricsClient),
			ListAssignedRoles:   query.NewListAssignedRolesQueryHandler(roleRepo, metricsClient),
			ListMenuCatalog:     query.NewListMenuCatalogQueryHandler(menuRepo, metricsClient),
			ListActions:         query.NewListActionsQueryHandler(permRepo, metricsClient),
			GetMenu:             query.NewGetMenuQueryHandler(menuRepo, metricsClient),
		},
	}
}
