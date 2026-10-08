package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type GetUserMenuTreeQuery struct {
	UserID uint64 `json:"user_id"`
}

func (q GetUserMenuTreeQuery) Validate() error {
	if q.UserID == 0 {
		return errors.New("invalid user id")
	}
	return nil
}

type GetUserMenuTreeQueryHandler decorator.QueryHandler[GetUserMenuTreeQuery, []*menu.Node]

type getUserMenuTreeQueryHandler struct {
	menuRepo menu.Repository
	permRepo permission.Repository
}

func NewGetUserMenuTreeQueryHandler(
	menuRepo menu.Repository,
	permRepo permission.Repository,
	metricsClient decorator.MetricsClient,
) GetUserMenuTreeQueryHandler {
	if menuRepo == nil || permRepo == nil {
		panic("nil RBAC Repository")
	}
	return decorator.ApplyQueryDecorators[GetUserMenuTreeQuery, []*menu.Node](
		getUserMenuTreeQueryHandler{menuRepo: menuRepo, permRepo: permRepo},
		metricsClient,
	)
}

func (h getUserMenuTreeQueryHandler) Handle(ctx context.Context, query GetUserMenuTreeQuery) ([]*menu.Node, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	matrix, err := h.permRepo.GetUserPermissionMatrix(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	menus, err := h.menuRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return menu.BuildTree(menus, matrix), nil
}
