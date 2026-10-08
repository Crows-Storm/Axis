package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type UpdateRoleMenuMaskCommand struct {
	RoleID uint64 `json:"role_id"`
	MenuID uint64 `json:"menu_id"`
	Mask   uint64 `json:"mask"`
	Grant  bool   `json:"grant"`
}

func (c UpdateRoleMenuMaskCommand) Validate() decorator.CommandError {
	if c.RoleID == 0 || c.MenuID == 0 {
		return decorator.InvalidCommand("role_id and menu_id are required")
	}
	return nil
}

type UpdateRoleMenuMaskCommandHandler decorator.CommandHandler[UpdateRoleMenuMaskCommand, struct{}]

type updateRoleMenuMaskCommandHandler struct {
	roleRepo role.Repository
	bus      decorator.EventBus
}

func NewUpdateRoleMenuMaskCommandHandler(
	roleRepo role.Repository,
	metricsClient decorator.MetricsClient,
	bus decorator.EventBus,
) UpdateRoleMenuMaskCommandHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyCommandDecorators[UpdateRoleMenuMaskCommand, struct{}](
		updateRoleMenuMaskCommandHandler{roleRepo: roleRepo, bus: bus},
		metricsClient,
	)
}

func (h updateRoleMenuMaskCommandHandler) Handle(ctx context.Context, cmd UpdateRoleMenuMaskCommand) (struct{}, error) {
	current, err := h.roleRepo.Get(ctx, cmd.RoleID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}

	mask := permission.Mask(cmd.Mask)
	if cmd.Grant {
		err = current.GrantMenu(cmd.MenuID, mask)
	} else {
		err = current.ReplaceMenuMask(cmd.MenuID, mask)
	}
	if err != nil {
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
