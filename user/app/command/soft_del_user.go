package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type SoftDeleteUserCommand struct {
	Id int64
}

func (s *SoftDeleteUserCommand) Validate() error {
	if s.Id <= 0 {
		return decorator.CommandExecutedError{Msg: "Invalid user ID"}
	}
	return nil
}

type SoftDeleteUserCommandHandler decorator.CommandHandler[SoftDeleteUserCommand, struct{}]

type softDeleteUserCommandHandler struct {
	userRepo domain.Repository
}

func NewSoftDeleteUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) SoftDeleteUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[SoftDeleteUserCommand, struct{}](
		softDeleteUserCommandHandler{userRepo: repo},
		metricsClient,
	)
}

// Handle implementation of `SoftDeleteUserCommand` returns void
func (c softDeleteUserCommandHandler) Handle(ctx context.Context, cmd SoftDeleteUserCommand) (struct{}, error) {
	err := c.userRepo.SoftDelete(ctx, cmd.Id)
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
