package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type GetUserRolesQuery struct {
	UserID uint64 `json:"user_id"`
}

func (q GetUserRolesQuery) Validate() error {
	if q.UserID == 0 {
		return errors.New("invalid user id")
	}
	return nil
}

type GetUserRolesQueryHandler decorator.QueryHandler[GetUserRolesQuery, []*role.Role]

type getUserRolesQueryHandler struct {
	roleRepo role.Repository
}

func NewGetUserRolesQueryHandler(
	roleRepo role.Repository,
	metricsClient decorator.MetricsClient,
) GetUserRolesQueryHandler {
	if roleRepo == nil {
		panic("nil Role Repository")
	}
	return decorator.ApplyQueryDecorators[GetUserRolesQuery, []*role.Role](
		getUserRolesQueryHandler{roleRepo: roleRepo},
		metricsClient,
	)
}

func (h getUserRolesQueryHandler) Handle(ctx context.Context, query GetUserRolesQuery) ([]*role.Role, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return h.roleRepo.GetByUserID(ctx, query.UserID)
}
