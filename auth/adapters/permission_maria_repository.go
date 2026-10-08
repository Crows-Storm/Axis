package adapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/server/store"
)

type RoleMenuModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	RoleID    uint64    `gorm:"column:role_id;not null;uniqueIndex:uk_role_menu,priority:1"`
	MenuID    uint64    `gorm:"column:menu_id;not null;uniqueIndex:uk_role_menu,priority:2"`
	Mask      uint64    `gorm:"column:mask;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (*RoleMenuModel) TableName() string {
	return "sys_role_menu"
}

type ActionModel struct {
	ID   uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Code string `gorm:"column:code;type:varchar(32);not null;uniqueIndex:uk_code"`
	Name string `gorm:"column:name;type:varchar(32);not null"`
	Bit  int8   `gorm:"column:bit;not null;uniqueIndex:uk_bit"`
	Sort int    `gorm:"column:sort;not null;default:0"`
}

func (*ActionModel) TableName() string {
	return "sys_action"
}

type PermissionMariaRepository struct {
	lock  *sync.RWMutex
	store *store.Store
}

func NewPermissionMariaRepository(st *store.Store) *PermissionMariaRepository {
	repo := &PermissionMariaRepository{
		lock:  &sync.RWMutex{},
		store: st,
	}
	if err := repo.autoMigrate(); err != nil {
		logger.WithError(err).Error("Failed to auto migrate permission tables")
	}
	return repo
}

func (r *PermissionMariaRepository) autoMigrate() error {
	return r.store.DB().AutoMigrate(&RoleMenuModel{}, &ActionModel{})
}

func (r *PermissionMariaRepository) GetUserMenuMask(ctx context.Context, userID uint64, menuCode string) (permission.Mask, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	type row struct {
		Mask uint64
	}
	var rows []row
	result := r.store.DB().WithContext(ctx).
		Table("sys_role_menu AS rm").
		Select("rm.mask").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Joins("JOIN sys_role r ON r.id = ur.role_id AND r.status = ?", 1).
		Joins("JOIN sys_menu m ON m.id = rm.menu_id AND m.status = ?", 1).
		Where("ur.user_id = ? AND m.code = ?", userID, menuCode).
		Scan(&rows)
	if result.Error != nil {
		return 0, fmt.Errorf("get user menu mask: %w", result.Error)
	}

	var mask permission.Mask
	for _, item := range rows {
		mask |= permission.Mask(item.Mask)
	}
	return mask, nil
}

func (r *PermissionMariaRepository) GetUserPermissionMatrix(ctx context.Context, userID uint64) (map[uint64]permission.Mask, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	type row struct {
		MenuID uint64
		Mask   uint64
	}
	var rows []row
	result := r.store.DB().WithContext(ctx).
		Table("sys_role_menu AS rm").
		Select("rm.menu_id, rm.mask").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Joins("JOIN sys_role r ON r.id = ur.role_id AND r.status = ?", 1).
		Joins("JOIN sys_menu m ON m.id = rm.menu_id AND m.status = ?", 1).
		Where("ur.user_id = ?", userID).
		Scan(&rows)
	if result.Error != nil {
		return nil, fmt.Errorf("get user permission matrix: %w", result.Error)
	}

	matrix := make(map[uint64]permission.Mask)
	for _, item := range rows {
		matrix[item.MenuID] |= permission.Mask(item.Mask)
	}
	return matrix, nil
}

func (r *PermissionMariaRepository) ListActions(ctx context.Context) ([]permission.ActionDef, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var models []ActionModel
	result := r.store.DB().WithContext(ctx).Order("sort asc, id asc").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("list actions: %w", result.Error)
	}
	if len(models) == 0 {
		return permission.BuiltinActions(), nil
	}
	out := make([]permission.ActionDef, 0, len(models))
	for _, m := range models {
		out = append(out, permission.ActionDef{
			ID:   m.ID,
			Code: m.Code,
			Name: m.Name,
			Bit:  int(m.Bit),
			Sort: m.Sort,
		})
	}
	return out, nil
}
