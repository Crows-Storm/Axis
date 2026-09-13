package command

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/common/decorator"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type CreateUserCommand struct {
	LoginID  string
	Password string
	Email    string
}

func (c CreateUserCommand) Validate() error {
	if c.LoginID == "" || c.Password == "" || c.Email == "" {
		return errors.New("invalid params")
	}
	return nil
}

type CreateUserCommandHandler decorator.CommandHandler[CreateUserCommand, struct{}]

type createUserCommandHandler struct {
	repo domain.Repository
}

func NewCreateUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) CreateUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[CreateUserCommand, struct{}](
		createUserCommandHandler{repo: repo},
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

	// TODO: handle in domain service
	//psw, err := utils.HashForStorage(cmd.Password)
	//if err != nil {
	//	logger.Errorf("hashing password error: %v", err)
	//	return struct{}{}, errors.New("invalid password")
	//}

	// Check if the command is valid
	if err := cmd.Validate(); err != nil {
		return struct{}{}, err
	}

	exist, err := c.repo.ExistsWithTransaction(ctx, 0, cmd.LoginID, "")
	if err != nil {
		return struct{}{}, err
	}
	if exist {
		return struct{}{}, commuser.ErrUserAlreadyExists
	}

	// factory fun to create/init aggregate root
	u, err := domain.Create(ctx, cmd.LoginID, cmd.Email)

	if err != nil {
		return struct{}{}, err
	}

	// apply password utils return the password to store
	if err := u.ApplyPassword(ctx, cmd.Password, false); err != nil {
		return struct{}{}, err
	}
	// At this point, two domain events will be generated

	// Domain Event Driver
	//	- An Event Store repository instance is required
	//	- The domain is built during the EvnetStore repository creation process
	//	- Within a transaction, the domain is inserted into `domain_events` and the event is pushed, awaiting ACK
	//	- ACK: Returns success if yes, failure if no
	//	- On the consumer side, the domain is inserted into the read model; returns yes on success, no on failure
	// TODO: caller event bus to publish events
	_, err = c.repo.Create(ctx, u) // to save
	if err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
