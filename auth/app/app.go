package app

import (
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/auth/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RegisterUser       command.RegisterUserCommandHandler
	Login              command.LoginCommandHandler
	Logout             command.LogoutCommandHandler
	AssignUserRole     command.AssignUserRoleCommandHandler
	UnassignUserRole   command.UnassignUserRoleCommandHandler
	ReplaceUserRoles   command.ReplaceUserRolesCommandHandler
	UpdateRoleMenuMask command.UpdateRoleMenuMaskCommandHandler
	RevokeRoleMenu     command.RevokeRoleMenuCommandHandler
	CreateRole         command.CreateRoleCommandHandler
	UpdateRole         command.UpdateRoleCommandHandler
	UpdateRoleStatus   command.UpdateRoleStatusCommandHandler
	CreateMenu         command.CreateMenuCommandHandler
	UpdateMenu         command.UpdateMenuCommandHandler
	UpdateMenuStatus   command.UpdateMenuStatusCommandHandler
}

type Queries struct {
	GetUser             query.GetUserQueryHandler
	GetUserRoles        query.GetUserRolesQueryHandler
	GetUserMenuTree     query.GetUserMenuTreeQueryHandler
	CheckMenuPermission query.CheckMenuPermissionQueryHandler
	ListRoles           query.ListRolesQueryHandler
	GetRole             query.GetRoleQueryHandler
	ListAssignedRoles   query.ListAssignedRolesQueryHandler
	ListMenuCatalog     query.ListMenuCatalogQueryHandler
	ListActions         query.ListActionsQueryHandler
	GetMenu             query.GetMenuQueryHandler
}
