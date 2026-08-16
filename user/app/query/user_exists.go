package query

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type UserExists struct {
	Id      int64
	LoginId string
	Email   string
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
	yes, err := g.userRepo.ExistsWithTransaction(ctx, query.Id, query.LoginId, query.Email)
	if err != nil {
		return false, err
	}
	return yes, err
}
