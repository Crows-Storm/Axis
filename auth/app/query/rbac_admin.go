package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type RoleView struct {
	ID          uint64 `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func toRoleView(r *role.Role) RoleView {
	if r == nil {
		return RoleView{}
	}
	return RoleView{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Status:      int8(r.Status),
	}
}

type RoleMenuView struct {
	MenuID  uint64   `json:"menu_id"`
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	Mask    uint64   `json:"mask"`
	Actions []string `json:"actions"`
}

type RoleDetail struct {
	RoleView
	Menus []RoleMenuView `json:"menus"`
}

type ListRolesQuery struct{}

type ListRolesQueryHandler decorator.QueryHandler[ListRolesQuery, []RoleView]

func NewListRolesQueryHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient) ListRolesQueryHandler {
	return decorator.ApplyQueryDecorators[ListRolesQuery, []RoleView](
		listRolesQueryHandler{roleRepo: roleRepo},
		metricsClient,
	)
}

type listRolesQueryHandler struct {
	roleRepo role.Repository
}

func (h listRolesQueryHandler) Handle(ctx context.Context, _ ListRolesQuery) ([]RoleView, error) {
	roles, err := h.roleRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleView(r))
	}
	return out, nil
}

type GetRoleQuery struct {
	RoleID uint64
}

func (q GetRoleQuery) Validate() error {
	if q.RoleID == 0 {
		return errors.New("invalid role id")
	}
	return nil
}

type GetRoleQueryHandler decorator.QueryHandler[GetRoleQuery, RoleDetail]

func NewGetRoleQueryHandler(roleRepo role.Repository, menuRepo menu.Repository, metricsClient decorator.MetricsClient) GetRoleQueryHandler {
	return decorator.ApplyQueryDecorators[GetRoleQuery, RoleDetail](
		getRoleQueryHandler{roleRepo: roleRepo, menuRepo: menuRepo},
		metricsClient,
	)
}

type getRoleQueryHandler struct {
	roleRepo role.Repository
	menuRepo menu.Repository
}

func (h getRoleQueryHandler) Handle(ctx context.Context, q GetRoleQuery) (RoleDetail, error) {
	if err := q.Validate(); err != nil {
		return RoleDetail{}, err
	}
	current, err := h.roleRepo.Get(ctx, q.RoleID)
	if err != nil {
		return RoleDetail{}, err
	}
	menus, err := h.menuRepo.ListAll(ctx)
	if err != nil {
		return RoleDetail{}, err
	}
	byID := make(map[uint64]*menu.Menu, len(menus))
	for _, m := range menus {
		byID[m.ID] = m
	}
	detail := RoleDetail{RoleView: toRoleView(current), Menus: []RoleMenuView{}}
	for _, grant := range current.Grants() {
		view := RoleMenuView{MenuID: grant.MenuID, Mask: uint64(grant.Mask), Actions: grant.Mask.Codes()}
		if m := byID[grant.MenuID]; m != nil {
			view.Code = m.Code
			view.Name = m.Name
		}
		detail.Menus = append(detail.Menus, view)
	}
	return detail, nil
}

type ListAssignedRolesQuery struct {
	UserID uint64
}

func (q ListAssignedRolesQuery) Validate() error {
	if q.UserID == 0 {
		return errors.New("invalid user id")
	}
	return nil
}

type ListAssignedRolesQueryHandler decorator.QueryHandler[ListAssignedRolesQuery, []RoleView]

func NewListAssignedRolesQueryHandler(roleRepo role.Repository, metricsClient decorator.MetricsClient) ListAssignedRolesQueryHandler {
	return decorator.ApplyQueryDecorators[ListAssignedRolesQuery, []RoleView](
		listAssignedRolesQueryHandler{roleRepo: roleRepo},
		metricsClient,
	)
}

type listAssignedRolesQueryHandler struct {
	roleRepo role.Repository
}

func (h listAssignedRolesQueryHandler) Handle(ctx context.Context, q ListAssignedRolesQuery) ([]RoleView, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	roles, err := h.roleRepo.ListAssignedByUserID(ctx, q.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleView(r))
	}
	return out, nil
}

type ListMenuCatalogQuery struct{}

type ListMenuCatalogQueryHandler decorator.QueryHandler[ListMenuCatalogQuery, []*menu.Node]

func NewListMenuCatalogQueryHandler(menuRepo menu.Repository, metricsClient decorator.MetricsClient) ListMenuCatalogQueryHandler {
	return decorator.ApplyQueryDecorators[ListMenuCatalogQuery, []*menu.Node](
		listMenuCatalogQueryHandler{menuRepo: menuRepo},
		metricsClient,
	)
}

type listMenuCatalogQueryHandler struct {
	menuRepo menu.Repository
}

func (h listMenuCatalogQueryHandler) Handle(ctx context.Context, _ ListMenuCatalogQuery) ([]*menu.Node, error) {
	menus, err := h.menuRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return menu.BuildCatalog(menus), nil
}

type ListActionsQuery struct{}

type ListActionsQueryHandler decorator.QueryHandler[ListActionsQuery, []permission.ActionDef]

func NewListActionsQueryHandler(permRepo permission.Repository, metricsClient decorator.MetricsClient) ListActionsQueryHandler {
	return decorator.ApplyQueryDecorators[ListActionsQuery, []permission.ActionDef](
		listActionsQueryHandler{permRepo: permRepo},
		metricsClient,
	)
}

type listActionsQueryHandler struct {
	permRepo permission.Repository
}

func (h listActionsQueryHandler) Handle(ctx context.Context, _ ListActionsQuery) ([]permission.ActionDef, error) {
	return h.permRepo.ListActions(ctx)
}

type GetMenuQuery struct {
	MenuID uint64
}

func (q GetMenuQuery) Validate() error {
	if q.MenuID == 0 {
		return errors.New("invalid menu id")
	}
	return nil
}

type GetMenuQueryHandler decorator.QueryHandler[GetMenuQuery, *menu.Menu]

func NewGetMenuQueryHandler(menuRepo menu.Repository, metricsClient decorator.MetricsClient) GetMenuQueryHandler {
	return decorator.ApplyQueryDecorators[GetMenuQuery, *menu.Menu](
		getMenuQueryHandler{menuRepo: menuRepo},
		metricsClient,
	)
}

type getMenuQueryHandler struct {
	menuRepo menu.Repository
}

func (h getMenuQueryHandler) Handle(ctx context.Context, q GetMenuQuery) (*menu.Menu, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	return h.menuRepo.Get(ctx, q.MenuID)
}
