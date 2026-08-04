package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type UpdateUserCommand struct {
	User      *domain.User
	UpdateFun func(context.Context, *domain.User) (*domain.User, error)
}

type UpdateUserCommandHandler decorator.CommandHandler[UpdateUserCommand, struct{}]

type updateUserCommandHandler struct {
	userRepo domain.Repository
}

func NewUpdateUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) UpdateUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[UpdateUserCommand, struct{}](
		updateUserCommandHandler{userRepo: repo},
		metricsClient,
	)

}

func (u updateUserCommandHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (struct{}, error) {
	if cmd.UpdateFun == nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: "The UpdateUserCommand command failed to Execute, because UpdateFun is Nil"}
	}
	err := u.userRepo.Update(ctx, cmd.User, cmd.UpdateFun)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: fmt.Sprintf("The UpdateUserCommand command failed to Execute, because: %v", err)}
	}

	return struct{}{}, nil
}
