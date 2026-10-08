package adapters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	domain "github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/server/store"
	"gorm.io/gorm"
)

type MenuModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	ParentID  uint64    `gorm:"column:parent_id;not null;default:0;index:idx_parent_id"`
	Code      string    `gorm:"column:code;type:varchar(64);not null;uniqueIndex:uk_code"`
	Name      string    `gorm:"column:name;type:varchar(64);not null"`
	Path      string    `gorm:"column:path;type:varchar(128)"`
	Icon      string    `gorm:"column:icon;type:varchar(64)"`
	Sort      int       `gorm:"column:sort;not null;default:0"`
	Status    int8      `gorm:"column:status;type:tinyint;default:1"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (*MenuModel) TableName() string {
	return "sys_menu"
}

func (m *MenuModel) toDomain() *domain.Menu {
	return domain.Reconstruct(m.ID, m.ParentID, m.Code, m.Name, m.Path, m.Icon, m.Sort, m.Status)
}

type MenuMariaRepository struct {
	lock  *sync.RWMutex
	store *store.Store
}

func NewMenuMariaRepository(st *store.Store) *MenuMariaRepository {
	repo := &MenuMariaRepository{
		lock:  &sync.RWMutex{},
		store: st,
	}
	if err := repo.autoMigrate(); err != nil {
		logger.WithError(err).Error("Failed to auto migrate menu table")
	}
	return repo
}

func (r *MenuMariaRepository) autoMigrate() error {
	return r.store.DB().AutoMigrate(&MenuModel{})
}

func (r *MenuMariaRepository) ListActive(ctx context.Context) ([]*domain.Menu, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var models []MenuModel
	result := r.store.DB().WithContext(ctx).
		Where("status = ?", 1).
		Order("sort asc, id asc").
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("list menus: %w", result.Error)
	}

	menus := make([]*domain.Menu, 0, len(models))
	for i := range models {
		menus = append(menus, models[i].toDomain())
	}
	return menus, nil
}

func (r *MenuMariaRepository) ListAll(ctx context.Context) ([]*domain.Menu, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var models []MenuModel
	result := r.store.DB().WithContext(ctx).Order("sort asc, id asc").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("list all menus: %w", result.Error)
	}
	menus := make([]*domain.Menu, 0, len(models))
	for i := range models {
		menus = append(menus, models[i].toDomain())
	}
	return menus, nil
}

func (r *MenuMariaRepository) Get(ctx context.Context, id uint64) (*domain.Menu, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var model MenuModel
	result := r.store.DB().WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMenuNotFound
		}
		return nil, fmt.Errorf("get menu: %w", result.Error)
	}
	return model.toDomain(), nil
}

func (r *MenuMariaRepository) Create(ctx context.Context, menu *domain.Menu) error {
	if menu == nil {
		return domain.ErrInvalidMenu
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	model := MenuModel{
		ParentID: menu.ParentID,
		Code:     menu.Code,
		Name:     menu.Name,
		Path:     menu.Path,
		Icon:     menu.Icon,
		Sort:     menu.Sort,
		Status:   menu.Status,
	}
	if err := r.store.DB().WithContext(ctx).Create(&model).Error; err != nil {
		if isUniqueConstraint(err) {
			return domain.ErrMenuCodeExist
		}
		return fmt.Errorf("create menu: %w", err)
	}
	menu.ID = model.ID
	return nil
}

func (r *MenuMariaRepository) Update(ctx context.Context, menu *domain.Menu) error {
	if menu == nil || menu.ID == 0 {
		return domain.ErrInvalidMenu
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	result := r.store.DB().WithContext(ctx).Model(&MenuModel{}).Where("id = ?", menu.ID).Updates(map[string]any{
		"parent_id": menu.ParentID,
		"name":      menu.Name,
		"path":      menu.Path,
		"icon":      menu.Icon,
		"sort":      menu.Sort,
		"status":    menu.Status,
	})
	if result.Error != nil {
		return fmt.Errorf("update menu: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrMenuNotFound
	}
	return nil
}

func (r *MenuMariaRepository) GetByCode(ctx context.Context, code string) (*domain.Menu, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var model MenuModel
	result := r.store.DB().WithContext(ctx).Where("code = ? AND status = ?", code, 1).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("menu not found: %s", code)
		}
		return nil, fmt.Errorf("get menu by code: %w", result.Error)
	}
	return model.toDomain(), nil
}
