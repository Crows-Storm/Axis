package command

import (
	"context"
	"strings"

	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
	commrole "github.com/Crows-Storm/Axis/common/domain/role"
)

type UpdateRoleCommand struct {
	RoleID      uint64 `json:"role_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c UpdateRoleCommand) Validate() decorator.CommandError {
	if c.RoleID == 0 || strings.TrimSpace(c.Name) == "" {
		return decorator.InvalidCommand("role_id and name are required")
	}
	return nil
}

type UpdateRoleCommandHandler decorator.CommandHandler[UpdateRoleCommand, struct{}]

func NewUpdateRoleCommandHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) UpdateRoleCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[UpdateRoleCommand, struct{}](
		updateRoleCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

type updateRoleCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func (h updateRoleCommandHandler) Handle(ctx context.Context, cmd UpdateRoleCommand) (struct{}, error) {
	current, err := h.roleRepo.Get(ctx, cmd.RoleID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.UpdateProfile(cmd.Name, cmd.Description); err != nil {
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

type UpdateRoleStatusCommand struct {
	RoleID uint64 `json:"role_id"`
	Status int8   `json:"status"`
}

func (c UpdateRoleStatusCommand) Validate() decorator.CommandError {
	if c.RoleID == 0 || (c.Status != 0 && c.Status != 1) {
		return decorator.InvalidCommand("role_id and status (0|1) are required")
	}
	return nil
}

type UpdateRoleStatusCommandHandler decorator.CommandHandler[UpdateRoleStatusCommand, struct{}]

func NewUpdateRoleStatusCommandHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) UpdateRoleStatusCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[UpdateRoleStatusCommand, struct{}](
		updateRoleStatusCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

type updateRoleStatusCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func (h updateRoleStatusCommandHandler) Handle(ctx context.Context, cmd UpdateRoleStatusCommand) (struct{}, error) {
	current, err := h.roleRepo.Get(ctx, cmd.RoleID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.ChangeStatus(commrole.Status(cmd.Status)); err != nil {
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
