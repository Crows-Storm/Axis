package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type DisableUserCommand struct {
	Id int64 `form:"id" binding:"required"`
}

type DisableUserCommandHandler decorator.CommandHandler[DisableUserCommand, struct{}]

type disableUserCommandHandler struct {
	userRepo domain.Repository
}

func NewDisableUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) DisableUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[DisableUserCommand, struct{}](
		disableUserCommandHandler{userRepo: repo},
		metricsClient,
	)
}

// Handle implementation of `DisableUserCommand` returns void
func (c disableUserCommandHandler) Handle(ctx context.Context, cmd DisableUserCommand) (struct{}, error) {
	// You can send domain events here to add other operations, such as:
	// - Create a user configuration record
	// - Send a welcome email (recorded in the task table)
	// - Create an audit log
	// If any operation fails, the entire transaction will be rolled back.
	err := c.userRepo.Disable(ctx, cmd.Id)
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
