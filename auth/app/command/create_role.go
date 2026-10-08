package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type CreateRoleCommand struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c CreateRoleCommand) Validate() decorator.CommandError {
	if c.Code == "" || c.Name == "" {
		return decorator.InvalidCommand("code and name are required")
	}
	return nil
}

type CreateRoleCommandHandler decorator.CommandHandler[CreateRoleCommand, *role.Role]

type createRoleCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func NewCreateRoleCommandHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) CreateRoleCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[CreateRoleCommand, *role.Role](
		createRoleCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

func (h createRoleCommandHandler) Handle(ctx context.Context, cmd CreateRoleCommand) (*role.Role, error) {
	created, err := role.Create(cmd.Code, cmd.Name, cmd.Description)
	if err != nil {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.roleRepo.Create(ctx, created); err != nil {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishRole(ctx, h.bus, created); err != nil {
		return nil, err
	}
	return created, nil
}
