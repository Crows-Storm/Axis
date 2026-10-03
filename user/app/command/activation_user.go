package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type ActivationUserCommand struct {
	ID uint64 `form:"id" binding:"required"`
}

type ActivationUserCommandHandler decorator.CommandHandler[ActivationUserCommand, struct{}]

type activationUserCommandHandler struct {
	repo domain.Repository
	bus  decorator.EventBus
}

func NewActivationUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
	bus decorator.EventBus,
) ActivationUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[ActivationUserCommand, struct{}](
		activationUserCommandHandler{repo: repo, bus: bus},
		metricsClient,
	)
}

// Handle implementation of `ActivationUserCommand` returns void
func (c activationUserCommandHandler) Handle(ctx context.Context, cmd ActivationUserCommand) (struct{}, error) {
	// You can send domain event here to add other operations, such as:
	// - Create a user configuration audit
	// - Send a welcome email (recorded in the task table)
	// - Create an audit log
	// If any operation fails, the entire transaction will be rolled back.

	user, err := c.repo.GetInfo(cmd.ID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "Invalid ID",
		}
	}
	if err := user.Activation(); err != nil {
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "Invalid ID",
		}
	}

	// call the repo
	if err := c.repo.StatusChange(ctx, user); err != nil {
		return struct{}{}, err
	}
	if err := publishUserEvents(ctx, c.bus, user); err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
