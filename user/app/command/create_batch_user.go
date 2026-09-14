package command

import (
	"context"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/util"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type CreateBatchUserCommand struct {
	Items []CreateUserItem
}

func (c CreateBatchUserCommand) Validate() error {
	if len(c.Items) == 0 {
		return fmt.Errorf("users list cannot be empty")
	}
	if len(c.Items) > 100 {
		return fmt.Errorf("cannot create more than 100 users at once")
	}

	// Check for duplicate LoginID and Email.
	LoginIDMap := make(map[string]bool)
	emailMap := make(map[string]bool)

	for _, user := range c.Items {
		if LoginIDMap[user.LoginID] {
			return fmt.Errorf("duplicate login_id: %s", user.LoginID)
		}
		if emailMap[user.Email] {
			return fmt.Errorf("duplicate email: %s", user.Email)
		}
		LoginIDMap[user.LoginID] = true
		emailMap[user.Email] = true
	}

	return nil
}

type CreateUserItem struct {
	LoginID  string
	Password string
	Email    string
}

type CreateBatchUserCommandHandler decorator.CommandHandler[CreateBatchUserCommand, struct{}]

type createBatchUserCommandHandler struct {
	userRepo domain.Repository
}

func NewCreateBatchUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) CreateBatchUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[CreateBatchUserCommand, struct{}](
		createBatchUserCommandHandler{userRepo: repo},
		metricsClient,
	)
}

// Handle implementation of `CreateBatchUserCommand` returns void
func (c createBatchUserCommandHandler) Handle(ctx context.Context, cmd CreateBatchUserCommand) (struct{}, error) {
	// Validate
	if err := cmd.Validate(); err != nil {
		return struct{}{}, err
	}
	users := make([]*domain.User, 0, len(cmd.Items))
	for _, item := range cmd.Items {
		users = append(users, &domain.User{
			ID:         util.GenerateID(),
			LoginID:    item.LoginID,
			Password:   item.Password,
			Email:      item.Email,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		})
	}

	err := c.userRepo.CreateBatch(ctx, users)
	if err != nil {
		return struct{}{}, err
	}

	// You can send domain events here to add other operations, such as:
	// - Create a user configuration record
	// - Send a welcome email (recorded in the task table)
	// - Create an audit log
	// If any operation fails, the entire transaction will be rolled back.
	return struct{}{}, nil
}
