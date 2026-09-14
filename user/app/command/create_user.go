package command

import (
	"context"
	"errors"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/decorator"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/util"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/Crows-Storm/Axis/user/utils"
)

type CreateUserCommand struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c CreateUserCommand) Validate() error {
	if c.LoginID == "" || c.Password == "" || c.Email == "" {
		return decorator.CommandExecutedError{Msg: decorator.InvalidCommand.Error()}
	}
	return nil
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

	psw, err := utils.HashForStorage(cmd.Password)
	if err != nil {
		logger.Errorf("hashing password error: %v", err)
		return struct{}{}, errors.New("invalid password")
	}
	exist, err := c.userRepo.ExistsWithTransaction(ctx, 0, cmd.LoginID, "")
	if err != nil {
		return struct{}{}, err
	}
	if exist {
		return struct{}{}, commuser.ErrUserAlreadyExists
	}
	_, err = c.userRepo.Create(ctx, &domain.User{
		ID:         util.GenerateID(),
		LoginID:    cmd.LoginID,
		Password:   psw, // is H1 + salt to storage
		Email:      cmd.Email,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	})
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
