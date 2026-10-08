package adapters

import (
	"context"
	"fmt"

	commrole "github.com/Crows-Storm/Axis/common/domain/role"
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server/store"
)

type MariaRBACSource struct {
	store *store.Store
}

func NewMariaRBACSource(st *store.Store) *MariaRBACSource {
	if st == nil {
		return nil
	}
	return &MariaRBACSource{store: st}
}

func (s *MariaRBACSource) EnabledRoleIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := s.store.DB().WithContext(ctx).
		Table("sys_role").
		Joins("JOIN sys_user_role ur ON ur.role_id = sys_role.id").
		Where("ur.user_id = ? AND sys_role.status = ?", userID, commrole.Active).
		Pluck("sys_role.id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("enabled role ids: %w", err)
	}
	return ids, nil
}

func (s *MariaRBACSource) RoleMenuMasks(ctx context.Context, roleID uint64) (map[string]uint64, error) {
	type row struct {
		Code string
		Mask uint64
	}
	var rows []row
	err := s.store.DB().WithContext(ctx).
		Table("sys_role_menu AS rm").
		Select("m.code, rm.mask").
		Joins("JOIN sys_role r ON r.id = rm.role_id AND r.status = ?", commrole.Active).
		Joins("JOIN sys_menu m ON m.id = rm.menu_id AND m.status = ?", 1).
		Where("rm.role_id = ?", roleID).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("role menu masks: %w", err)
	}
	out := make(map[string]uint64, len(rows))
	for _, item := range rows {
		if item.Code == "" {
			continue
		}
		out[item.Code] |= item.Mask
	}
	return out, nil
}

func (s *MariaRBACSource) RoleIDsForMenuCode(ctx context.Context, menuCode string) ([]uint64, error) {
	var ids []uint64
	err := s.store.DB().WithContext(ctx).
		Table("sys_role_menu AS rm").
		Joins("JOIN sys_menu m ON m.id = rm.menu_id").
		Where("m.code = ?", menuCode).
		Pluck("rm.role_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("role ids for menu: %w", err)
	}
	return ids, nil
}

var _ rbac.Source = (*MariaRBACSource)(nil)
