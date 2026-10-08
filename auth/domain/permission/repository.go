package permission

import (
	"context"
	"errors"
)

var (
	ErrNoPermission = errors.New("no permission")
	ErrMenuNotFound = errors.New("menu not found")
)

// Repository is the query-side port for merged user permission snapshots.
type ActionDef struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Bit  int    `json:"bit"`
	Sort int    `json:"sort"`
}

func BuiltinActions() []ActionDef {
	out := make([]ActionDef, 0, len(ActionCode))
	for a, code := range ActionCode {
		out = append(out, ActionDef{
			ID:   uint64(a) + 1,
			Code: code,
			Name: ActionName[a],
			Bit:  int(a),
			Sort: int(a) + 1,
		})
	}
	return out
}

type Repository interface {
	GetUserMenuMask(ctx context.Context, userID uint64, menuCode string) (Mask, error)
	GetUserPermissionMatrix(ctx context.Context, userID uint64) (map[uint64]Mask, error)
	ListActions(ctx context.Context) ([]ActionDef, error)
}
