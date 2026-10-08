package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type UnassignUserRoleCommand struct {
	UserID uint64 `json:"user_id"`
	RoleID uint64 `json:"role_id"`
}

func (c UnassignUserRoleCommand) Validate() decorator.CommandError {
	if c.UserID == 0 || c.RoleID == 0 {
		return decorator.InvalidCommand("user_id and role_id are required")
	}
	return nil
}

type UnassignUserRoleCommandHandler decorator.CommandHandler[UnassignUserRoleCommand, struct{}]

func NewUnassignUserRoleCommandHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) UnassignUserRoleCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[UnassignUserRoleCommand, struct{}](
		unassignUserRoleCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

type unassignUserRoleCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func (h unassignUserRoleCommandHandler) Handle(ctx context.Context, cmd UnassignUserRoleCommand) (struct{}, error) {
	current, err := h.roleRepo.Get(ctx, cmd.RoleID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.UnassignFrom(cmd.UserID); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.roleRepo.UnassignUser(ctx, cmd.UserID, current); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishRole(ctx, h.bus, current); err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}

type ReplaceUserRolesCommand struct {
	UserID  uint64   `json:"user_id"`
	RoleIDs []uint64 `json:"role_ids"`
}

func (c ReplaceUserRolesCommand) Validate() decorator.CommandError {
	if c.UserID == 0 {
		return decorator.InvalidCommand("user_id is required")
	}
	return nil
}

type ReplaceUserRolesCommandHandler decorator.CommandHandler[ReplaceUserRolesCommand, struct{}]

func NewReplaceUserRolesCommandHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) ReplaceUserRolesCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[ReplaceUserRolesCommand, struct{}](
		replaceUserRolesCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

type replaceUserRolesCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func (h replaceUserRolesCommandHandler) Handle(ctx context.Context, cmd ReplaceUserRolesCommand) (struct{}, error) {
	roles := make([]*role.Role, 0, len(cmd.RoleIDs))
	for _, id := range cmd.RoleIDs {
		current, err := h.roleRepo.Get(ctx, id)
		if err != nil {
			return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
		}
		if err := current.AssignTo(cmd.UserID); err != nil {
			return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
		}
		roles = append(roles, current)
	}
	if err := h.roleRepo.ReplaceUserRoles(ctx, cmd.UserID, roles); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	for _, current := range roles {
		if err := publishRole(ctx, h.bus, current); err != nil {
			return struct{}{}, err
		}
	}
	return struct{}{}, nil
}

type RevokeRoleMenuCommand struct {
	RoleID uint64 `json:"role_id"`
	MenuID uint64 `json:"menu_id"`
}

func (c RevokeRoleMenuCommand) Validate() decorator.CommandError {
	if c.RoleID == 0 || c.MenuID == 0 {
		return decorator.InvalidCommand("role_id and menu_id are required")
	}
	return nil
}

type RevokeRoleMenuCommandHandler decorator.CommandHandler[RevokeRoleMenuCommand, struct{}]

func NewRevokeRoleMenuCommandHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) RevokeRoleMenuCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[RevokeRoleMenuCommand, struct{}](
		revokeRoleMenuCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

type revokeRoleMenuCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func (h revokeRoleMenuCommandHandler) Handle(ctx context.Context, cmd RevokeRoleMenuCommand) (struct{}, error) {
	current, err := h.roleRepo.Get(ctx, cmd.RoleID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.RevokeMenu(cmd.MenuID); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	err = h.roleRepo.Update(ctx, current, func(_ context.Context, _ *role.Role) (*role.Role, error) {
		return current, nil
	})
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishRole(ctx, h.bus, current); err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
