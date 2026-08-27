package query

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
)

type GetUserQuery struct {
	Id      int64
	LoginId string
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
	if query.LoginId != "" {
		info, err = g.userRepo.GetByLoginId(ctx, query.LoginId)
	} else if query.Id > 0 {
		info, err = g.userRepo.GetInfo(query.Id)
	}
	if err != nil {
		return nil, err
	}
	return info, err
}
