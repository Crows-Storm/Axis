package command

import (
	"context"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
)

type CreateUserCommand struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c CreateUserCommand) Validate() decorator.CommandError {
	if c.LoginID == "" || c.Password == "" || c.Email == "" {
		return decorator.InvalidCommand("login id or password or email cannot be empty.")
	}
	return nil
}

type CreateUserCommandHandler decorator.CommandHandler[CreateUserCommand, struct{}]

type createUserCommandHandler struct {
	repo   domain.Repository
	hasher domain.PasswordHasher
}

func NewCreateUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
	hasher domain.PasswordHasher,
) CreateUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[CreateUserCommand, struct{}](
		createUserCommandHandler{repo: repo, hasher: hasher},
		metricsClient,
	)
}

// Handle implementation of `CreateUserCommand` returns void
func (c createUserCommandHandler) Handle(ctx context.Context, cmd CreateUserCommand) (struct{}, error) {
	exist, err := c.repo.ExistsWithTransaction(ctx, 0, cmd.LoginID, "")
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "Invalid LoginID",
		}
	}
	//r, err := errgroup.WithContext(ctx)
	//r.Go(func() error {
	//	// do something
	//	return nil
	//})

	if exist {
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "user already exists",
		}
	}

	// factory fun to create/init aggregate root
	u, err := domain.Create(cmd.LoginID, cmd.Email)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"LoginID": cmd.LoginID,
			"Email":   cmd.Email,
		}).WithError(err).Error("failed to create user")
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "Invalid login id or email",
		}
	}

	// apply password utils return the password to store
	if err := u.ApplyPassword(cmd.Password, false, c.hasher); err != nil {

		logger.WithFields(logrus.Fields{
			"LoginID":  cmd.LoginID,
			"Email":    cmd.Email,
			"Password": cmd.Password,
		}).Warn("Failed to apply password when creating user")

		return struct{}{}, decorator.CommandExecutedError{
			Msg: "Invalid login id or password",
		}
	}

	_, err = c.repo.Create(ctx, u) // to save
	if err != nil {
		logger.WithFields(logrus.Fields{
			"LoginID": cmd.LoginID,
			"Email":   cmd.Email,
		}).Warn("failed to create user")
		return struct{}{}, decorator.CommandExecutedError{
			Msg: "Invalid login id or password",
		}
	}

	// TODO: do Publish Event on bus

	return struct{}{}, nil
}
