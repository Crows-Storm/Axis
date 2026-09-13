package command

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/decorator"
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

	// Check for duplicate LoginId and Email.
	loginIdMap := make(map[string]bool)
	emailMap := make(map[string]bool)

	for _, user := range c.Items {
		if loginIdMap[user.LoginId] {
			return fmt.Errorf("duplicate login_id: %s", user.LoginId)
		}
		if emailMap[user.Email] {
			return fmt.Errorf("duplicate email: %s", user.Email)
		}
		loginIdMap[user.LoginId] = true
		emailMap[user.Email] = true
	}

	return nil
}

type CreateUserItem struct {
	LoginId  string
	Password string
	Email    string
}

type CreateBatchUserCommandHandler decorator.CommandHandler[CreateBatchUserCommand, struct{}]

type createBatchUserCommandHandler struct {
	repo domain.Repository
}

func NewCreateBatchUserCommandHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) CreateBatchUserCommandHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyCommandDecorators[CreateBatchUserCommand, struct{}](
		createBatchUserCommandHandler{repo: repo},
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
			//Id:         util.GenerateFlakeID(), need use domain create
			LoginId: item.LoginId,
			Email:   item.Email,
		})
	}

	err := c.repo.CreateBatch(ctx, users)
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
