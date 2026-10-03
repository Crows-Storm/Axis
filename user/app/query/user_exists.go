package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type UserExists struct {
	ID      uint64 `json:"id" form:"id"`
	LoginID string `json:"login_id" form:"login_id"`
	Email   string `json:"email" form:"email"`
}

func (u UserExists) Validate() decorator.CommandError {
	count := 0
	if u.ID != 0 {
		count++
	}
	if u.LoginID != "" {
		count++
	}
	if u.Email != "" {
		count++
	}
	if count != 1 {
		return errors.New("exactly one of ID, LoginID, Email must be set")
	}
	return nil
}

type UserExistsHandler decorator.QueryHandler[UserExists, bool]

type userExistsHandler struct {
	userRepo domain.Repository
}

func NewUserExistsHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) UserExistsHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyQueryDecorators[UserExists, bool](
		userExistsHandler{userRepo: repo},
		metricsClient,
	)
}

func (g userExistsHandler) Handle(ctx context.Context, query UserExists) (bool, error) {
	yes, err := g.userRepo.ExistsWithTransaction(ctx, query.ID, query.LoginID, query.Email)
	if err != nil {
		return false, err
	}
	return yes, err
}
