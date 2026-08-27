package adapters

import (
	"context"
	"sync"
	"time"

	domain "github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/server/store"
)

type RoleModel struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	RoleCode   string    `gorm:"column:role_code;type:varchar(50);not null;uniqueIndex:uk_role_code_deleted,priority:1"`
	RoleName   string    `gorm:"column:role_name;type:varchar(50);not null"`
	RoleLevel  int       `gorm:"column:role_level;type:int;default:0"`
	Status     int8      `gorm:"column:status;type:tinyint;default:1;index:idx_status"`
	Deleted    int8      `gorm:"column:deleted;type:tinyint;default:0;uniqueIndex:uk_role_code_deleted,priority:2"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (*RoleModel) TableName() string {
	return "sys_role"
}

type RoleMariaRepository struct {
	lock  *sync.RWMutex
	store *store.Store
}

func NewRoleMariaRepository(store *store.Store) *RoleMariaRepository {
	repo := &RoleMariaRepository{
		lock:  &sync.RWMutex{},
		store: store,
	}

	if err := repo.autoMigrate(); err != nil {
		logger.WithError(err).Error("Failed to auto migrate role table")
	}

	return repo
}

func (r *RoleMariaRepository) autoMigrate() error {
	return r.store.DB().AutoMigrate(&RoleModel{})
}

func (r *RoleMariaRepository) toDomainList(models []*RoleModel) []*domain.Role {
	roles := make([]*domain.Role, 0, len(models))
	//for _, model := range models {
	//	roles = append(roles, r.toDomain(model))
	//}
	return roles
}

func (u *RoleMariaRepository) Get(ctx context.Context, id int64) (*domain.Role, error) {
	//TODO implement me
	panic("implement me")
}

func (u *RoleMariaRepository) GetByUserId(ctx context.Context, userId int64) (*domain.Role, error) {
	//TODO implement me
	panic("implement me")
}
