package command

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/util"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
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
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CreateUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[CreateUserCommand, struct{}](
		createUserCommandHandler{userRepo: repo},
		logger,
		metricsClient,
	)
}

// Handle implementation of `CreateUserCommand` returns void
func (c createUserCommandHandler) Handle(ctx context.Context, cmd CreateUserCommand) (struct{}, error) {
	_, err := c.userRepo.Create(ctx, &domain.User{
		Id:         util.GenerateID(),
		LoginId:    cmd.LoginId,
		Password:   cmd.Password,
		Email:      cmd.Email,
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	})
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
