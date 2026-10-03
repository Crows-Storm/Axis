package adapters

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	domain "github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/config/logger"
	commrole "github.com/Crows-Storm/Axis/common/domain/role"
	"github.com/Crows-Storm/Axis/common/server/store"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleModel struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Code        string    `gorm:"column:code;type:varchar(64);not null;uniqueIndex:uk_code"`
	Name        string    `gorm:"column:name;type:varchar(64);not null"`
	Description string    `gorm:"column:description;type:varchar(255)"`
	Status      int8      `gorm:"column:status;type:tinyint;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (*RoleModel) TableName() string {
	return "sys_role"
}

type UserRoleModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_user_role,priority:1"`
	RoleID    uint64    `gorm:"column:role_id;not null;uniqueIndex:uk_user_role,priority:2"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (*UserRoleModel) TableName() string {
	return "sys_user_role"
}

type roleGrantModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	RoleID    uint64    `gorm:"column:role_id;not null;uniqueIndex:uk_role_menu,priority:1"`
	MenuID    uint64    `gorm:"column:menu_id;not null;uniqueIndex:uk_role_menu,priority:2"`
	Mask      uint64    `gorm:"column:mask;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (*roleGrantModel) TableName() string {
	return "sys_role_menu"
}

type RoleMariaRepository struct {
	lock  *sync.RWMutex
	store *store.Store
}

func NewRoleMariaRepository(st *store.Store) *RoleMariaRepository {
	repo := &RoleMariaRepository{
		lock:  &sync.RWMutex{},
		store: st,
	}
	if err := repo.autoMigrate(); err != nil {
		logger.WithError(err).Error("Failed to auto migrate role tables")
	}
	return repo
}

func (r *RoleMariaRepository) autoMigrate() error {
	return r.store.DB().AutoMigrate(&RoleModel{}, &UserRoleModel{}, &roleGrantModel{})
}

func (r *RoleMariaRepository) Get(ctx context.Context, id uint64) (*domain.Role, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.getUnlocked(ctx, id)
}

func (r *RoleMariaRepository) getUnlocked(ctx context.Context, id uint64) (*domain.Role, error) {
	var model RoleModel
	result := r.store.DB().WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("get role: %w", result.Error)
	}

	grants, err := r.loadGrantsUnlocked(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return domain.Reconstruct(
		model.ID,
		model.Code,
		model.Name,
		model.Description,
		commrole.Status(model.Status),
		grants,
	), nil
}

func (r *RoleMariaRepository) loadGrantsUnlocked(ctx context.Context, roleID uint64) (map[uint64]permission.Mask, error) {
	var models []roleGrantModel
	result := r.store.DB().WithContext(ctx).Where("role_id = ?", roleID).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("load role grants: %w", result.Error)
	}
	grants := make(map[uint64]permission.Mask, len(models))
	for _, m := range models {
		grants[m.MenuID] = permission.Mask(m.Mask)
	}
	return grants, nil
}

func (r *RoleMariaRepository) GetByUserID(ctx context.Context, userID uint64) ([]*domain.Role, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var models []RoleModel
	result := r.store.DB().WithContext(ctx).
		Table("sys_role").
		Joins("JOIN sys_user_role ur ON ur.role_id = sys_role.id").
		Where("ur.user_id = ? AND sys_role.status = ?", userID, commrole.Active).
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("get roles by user: %w", result.Error)
	}

	roles := make([]*domain.Role, 0, len(models))
	for i := range models {
		grants, err := r.loadGrantsUnlocked(ctx, models[i].ID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, domain.Reconstruct(
			models[i].ID,
			models[i].Code,
			models[i].Name,
			models[i].Description,
			commrole.Status(models[i].Status),
			grants,
		))
	}
	return roles, nil
}

func (r *RoleMariaRepository) List(ctx context.Context) ([]*domain.Role, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var models []RoleModel
	result := r.store.DB().WithContext(ctx).Order("id asc").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("list roles: %w", result.Error)
	}
	roles := make([]*domain.Role, 0, len(models))
	for i := range models {
		roles = append(roles, domain.Reconstruct(
			models[i].ID,
			models[i].Code,
			models[i].Name,
			models[i].Description,
			commrole.Status(models[i].Status),
			nil,
		))
	}
	return roles, nil
}

func (r *RoleMariaRepository) ListAssignedByUserID(ctx context.Context, userID uint64) ([]*domain.Role, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var models []RoleModel
	result := r.store.DB().WithContext(ctx).
		Table("sys_role").
		Joins("JOIN sys_user_role ur ON ur.role_id = sys_role.id").
		Where("ur.user_id = ?", userID).
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("list assigned roles: %w", result.Error)
	}
	roles := make([]*domain.Role, 0, len(models))
	for i := range models {
		roles = append(roles, domain.Reconstruct(
			models[i].ID,
			models[i].Code,
			models[i].Name,
			models[i].Description,
			commrole.Status(models[i].Status),
			nil,
		))
	}
	return roles, nil
}

func (r *RoleMariaRepository) Create(ctx context.Context, role *domain.Role) error {
	if role == nil {
		return domain.ErrInvalidRole
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	model := RoleModel{
		ID:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		Status:      int8(role.Status),
	}
	if err := r.store.DB().WithContext(ctx).Create(&model).Error; err != nil {
		if isUniqueConstraint(err) {
			return domain.ErrRoleCodeExist
		}
		return fmt.Errorf("create role: %w", err)
	}
	role.ID = model.ID
	return nil
}

func (r *RoleMariaRepository) AssignUser(ctx context.Context, userID uint64, role *domain.Role) error {
	if role == nil {
		return domain.ErrInvalidRole
	}
	if userID == 0 {
		return domain.ErrInvalidAssignment
	}
	if !role.IsActive() {
		return domain.ErrRoleDisabled
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	var persisted RoleModel
	if err := r.store.DB().WithContext(ctx).Where("id = ?", role.ID).First(&persisted).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrRoleNotFound
		}
		return fmt.Errorf("assign user role: %w", err)
	}

	model := UserRoleModel{UserID: userID, RoleID: role.ID}
	result := r.store.DB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
		DoNothing: true,
	}).Create(&model)
	if result.Error != nil {
		return fmt.Errorf("assign user role: %w", result.Error)
	}
	return nil
}

func (r *RoleMariaRepository) UnassignUser(ctx context.Context, userID uint64, role *domain.Role) error {
	if role == nil {
		return domain.ErrInvalidRole
	}
	if userID == 0 {
		return domain.ErrInvalidAssignment
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	result := r.store.DB().WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, role.ID).
		Delete(&UserRoleModel{})
	if result.Error != nil {
		return fmt.Errorf("unassign user role: %w", result.Error)
	}
	return nil
}

func (r *RoleMariaRepository) ReplaceUserRoles(ctx context.Context, userID uint64, roles []*domain.Role) error {
	if userID == 0 {
		return domain.ErrInvalidAssignment
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.store.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&UserRoleModel{}).Error; err != nil {
			return fmt.Errorf("clear user roles: %w", err)
		}
		for _, role := range roles {
			if role == nil {
				return domain.ErrInvalidRole
			}
			if !role.IsActive() {
				return domain.ErrRoleDisabled
			}
			model := UserRoleModel{UserID: userID, RoleID: role.ID}
			if err := tx.Create(&model).Error; err != nil {
				return fmt.Errorf("replace user roles: %w", err)
			}
		}
		return nil
	})
}

func (r *RoleMariaRepository) Update(
	ctx context.Context,
	role *domain.Role,
	updateFun func(context.Context, *domain.Role) (*domain.Role, error),
) error {
	if role == nil {
		return domain.ErrInvalidRole
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	current, err := r.getUnlocked(ctx, role.ID)
	if err != nil {
		return err
	}

	updated, err := updateFun(ctx, current)
	if err != nil {
		return err
	}
	if updated == nil {
		return domain.ErrInvalidRole
	}

	return r.persistUnlocked(ctx, updated)
}

func (r *RoleMariaRepository) persistUnlocked(ctx context.Context, role *domain.Role) error {
	model := RoleModel{
		ID:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		Status:      int8(role.Status),
	}
	if err := r.store.DB().WithContext(ctx).Model(&RoleModel{}).Where("id = ?", role.ID).Updates(map[string]any{
		"name":        model.Name,
		"description": model.Description,
		"status":      model.Status,
	}).Error; err != nil {
		return fmt.Errorf("save role: %w", err)
	}
	return r.saveGrantsUnlocked(ctx, role)
}

func (r *RoleMariaRepository) saveGrantsUnlocked(ctx context.Context, role *domain.Role) error {
	keep := make([]uint64, 0, len(role.Grants()))
	for _, grant := range role.Grants() {
		keep = append(keep, grant.MenuID)
		model := roleGrantModel{
			RoleID: role.ID,
			MenuID: grant.MenuID,
			Mask:   uint64(grant.Mask),
		}
		result := r.store.DB().WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "role_id"}, {Name: "menu_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"mask"}),
		}).Create(&model)
		if result.Error != nil {
			return fmt.Errorf("save role grants: %w", result.Error)
		}
	}
	q := r.store.DB().WithContext(ctx).Where("role_id = ?", role.ID)
	if len(keep) == 0 {
		if err := q.Delete(&roleGrantModel{}).Error; err != nil {
			return fmt.Errorf("clear role grants: %w", err)
		}
		return nil
	}
	if err := q.Where("menu_id NOT IN ?", keep).Delete(&roleGrantModel{}).Error; err != nil {
		return fmt.Errorf("prune role grants: %w", err)
	}
	return nil
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}
