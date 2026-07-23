package query

import (
	"context"

	"github.com/Crows-Storm/Axis/common/decorator"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
)

type GetUserQuery struct {
	Id int64
}

type GetUserQueryHandler decorator.QueryHandler[GetUserQuery, *domain.User]

type getUserQueryHandler struct {
	userRepo domain.Repository
}

func NewGetUserQueryHandler(
	repo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) GetUserQueryHandler {
	if repo == nil {
		panic("nil User Repository")
	}
	return decorator.ApplyQueryDecorators[GetUserQuery, *domain.User](
		getUserQueryHandler{userRepo: repo},
		logger,
		metricsClient,
	)
}

func (g getUserQueryHandler) Handle(ctx context.Context, query GetUserQuery) (*domain.User, error) {
	info, err := g.userRepo.GetInfo(query.Id)
	if err != nil {
		return nil, err
	}
	return info, err
}
