package query

import (
	"context"
	"errors"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/rbac"
)

type CheckMenuPermissionQuery struct {
	UserID   uint64 `json:"user_id"`
	MenuCode string `json:"menu_code"`
	Action   string `json:"action"`
}

func (q CheckMenuPermissionQuery) Validate() error {
	if q.UserID == 0 || q.MenuCode == "" || q.Action == "" {
		return errors.New("user_id, menu_code and action are required")
	}
	if _, ok := permission.ParseAction(q.Action); !ok {
		return errors.New("unknown action")
	}
	return nil
}

type CheckMenuPermissionResult struct {
	Allowed bool     `json:"allowed"`
	Mask    uint64   `json:"mask"`
	Actions []string `json:"actions"`
}

type CheckMenuPermissionQueryHandler decorator.QueryHandler[CheckMenuPermissionQuery, CheckMenuPermissionResult]

type checkMenuPermissionQueryHandler struct {
	authz *rbac.Authorizer
}

func NewCheckMenuPermissionQueryHandler(
	authz *rbac.Authorizer,
	metricsClient decorator.MetricsClient,
) CheckMenuPermissionQueryHandler {
	if authz == nil {
		panic("nil RBAC authorizer")
	}
	return decorator.ApplyQueryDecorators[CheckMenuPermissionQuery, CheckMenuPermissionResult](
		checkMenuPermissionQueryHandler{authz: authz},
		metricsClient,
	)
}

func (h checkMenuPermissionQueryHandler) Handle(ctx context.Context, query CheckMenuPermissionQuery) (CheckMenuPermissionResult, error) {
	if err := query.Validate(); err != nil {
		return CheckMenuPermissionResult{}, err
	}
	action, _ := permission.ParseAction(query.Action)
	mask, err := h.authz.UserMenuMask(ctx, query.UserID, query.MenuCode)
	if err != nil {
		return CheckMenuPermissionResult{}, err
	}
	return CheckMenuPermissionResult{
		Allowed: mask.Has(rbac.Action(action)),
		Mask:    uint64(mask),
		Actions: mask.Codes(),
	}, nil
}
