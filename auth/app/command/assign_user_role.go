package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type AssignUserRoleCommand struct {
	UserID uint64 `json:"user_id"`
	RoleID uint64 `json:"role_id"`
}

func (c AssignUserRoleCommand) Validate() decorator.CommandError {
	if c.UserID == 0 || c.RoleID == 0 {
		return decorator.InvalidCommand("user_id and role_id are required")
	}
	return nil
}

type AssignUserRoleCommandHandler decorator.CommandHandler[AssignUserRoleCommand, struct{}]

type assignUserRoleCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func NewAssignUserRoleCommandHandler(
	roleRepo role.Repository,
	metricsClient decorator.MetricsClient,
	bus decorator.EventBus,
) AssignUserRoleCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[AssignUserRoleCommand, struct{}](
		assignUserRoleCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

func (h assignUserRoleCommandHandler) Handle(ctx context.Context, cmd AssignUserRoleCommand) (struct{}, error) {
	current, err := h.roleRepo.Get(ctx, cmd.RoleID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.AssignTo(cmd.UserID); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.roleRepo.AssignUser(ctx, cmd.UserID, current); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishRole(ctx, h.bus, current); err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
