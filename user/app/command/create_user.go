package command

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/util"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type CreateUserCommand struct {
	LoginId  string
	Password string
	Email    string
}

type CreateUserCommandHandler decorator.CommandHandler[CreateUserCommand, struct{}]

type createUserCommandHandler struct {
	userRepo domain.Repository
}

func NewCreateUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) CreateUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[CreateUserCommand, struct{}](
		createUserCommandHandler{userRepo: repo},
		metricsClient,
	)
}

// Handle implementation of `CreateUserCommand` returns void
func (c createUserCommandHandler) Handle(ctx context.Context, cmd CreateUserCommand) (struct{}, error) {
	// You can send domain events here to add other operations, such as:
	// - Create a user configuration record
	// - Send a welcome email (recorded in the task table)
	// - Create an audit log
	// If any operation fails, the entire transaction will be rolled back.
	_, err := c.userRepo.Create(ctx, &domain.User{
		Id:         util.GenerateID(),
		LoginId:    cmd.LoginId,
		Password:   cmd.Password,
		Email:      cmd.Email,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	})
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
