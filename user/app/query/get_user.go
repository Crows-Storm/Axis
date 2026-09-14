package query

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type GetUserQuery struct {
	ID      uint64 `json:"id"`
	LoginID string `json:"login_id"`
}

type GetUserQueryHandler decorator.QueryHandler[GetUserQuery, *domain.User]

type getUserQueryHandler struct {
	userRepo domain.Repository
}

func NewGetUserQueryHandler(
	repo domain.Repository,
	metricsClient decorator.MetricsClient,
) GetUserQueryHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyQueryDecorators[GetUserQuery, *domain.User](
		getUserQueryHandler{userRepo: repo},
		metricsClient,
	)
}

func (g getUserQueryHandler) Handle(ctx context.Context, query GetUserQuery) (*domain.User, error) {
	var info *domain.User
	var err error
	if query.LoginID != "" {
		info, err = g.userRepo.GetByLoginID(ctx, query.LoginID)
	} else if query.ID > 0 {
		info, err = g.userRepo.GetInfo(query.ID)
	}
	if err != nil {
		return nil, err
	}
	return info, err
}
