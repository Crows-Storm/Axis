package adapters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/server/store"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModel struct {
	Id         int64     `gorm:"column:id;unsigned;primaryKey;"`
	LoginId    string    `gorm:"column:login_id;uniqueIndex;type:varchar(50);not null"`
	Password   string    `gorm:"column:password;type:varchar(255);not null"`
	Email      string    `gorm:"column:email;uniqueIndex;type:varchar(100);not null"`
	Status     int8      `gorm:"column:status;type:tinyint;default:1;not null;comment:'1:normal 0:disable'"`
	Deleted    int8      `gorm:"column:deleted;type:tinyint;default:0;not null;index;comment:'0:not deleted 1:deleted'"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null;autoUpdateTime"`
}

func (*UserModel) TableName() string {
	return "sys_user"
}

func (m *UserModel) toDomain() *domain.User {
	return &domain.User{
		Id:         m.Id,
		LoginId:    m.LoginId,
		Password:   m.Password,
		Email:      m.Email,
		Status:     m.Status,
		Deleted:    m.Deleted,
		CreateTime: m.CreateTime,
		UpdateTime: m.UpdateTime,
	}
}

func fromDomain(user *domain.User) *UserModel {
	return &UserModel{
		Id:         user.Id,
		LoginId:    user.LoginId,
		Password:   user.Password,
		Email:      user.Email,
		Status:     user.Status,
		Deleted:    user.Deleted,
		CreateTime: user.CreateTime,
		UpdateTime: user.UpdateTime,
	}
}

type UserMariaRepository struct {
	lock  *sync.RWMutex
	store *store.Store
}

func NewUserMariaRepository(store *store.Store) *UserMariaRepository {
	repo := &UserMariaRepository{
		lock:  &sync.RWMutex{},
		store: store,
	}

	if err := repo.autoMigrate(); err != nil {
		logger.WithError(err).Error("Failed to auto migrate user table")
	}

	return repo
}

func (u *UserMariaRepository) autoMigrate() error {
	return u.store.DB().AutoMigrate(&UserModel{})
}

func (u *UserMariaRepository) GetInfo(id int64) (*domain.User, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	var userModel UserModel
	result := u.store.DB().
		Where("id = ? AND deleted = 0", id).
		First(&userModel)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.NotFoundError{UserId: id}
		}
		logger.WithError(result.Error).WithField("user_id", id).Error("Failed to get user by id")
		return nil, fmt.Errorf("failed to get user by id: %w", result.Error)
	}

	return userModel.toDomain(), nil
}

func (u *UserMariaRepository) GetByLoginId(ctx context.Context, loginId string) (*domain.User, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	var userModel UserModel
	result := u.store.DB().WithContext(ctx).
		Where("login_id = ? AND deleted = 0", loginId).
		First(&userModel)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", loginId)
		}
		logger.WithError(result.Error).WithField("login_id", loginId).Error("Failed to get user by login_id")
		return nil, fmt.Errorf("failed to get user by login_id: %w", result.Error)
	}

	return userModel.toDomain(), nil
}

func (u *UserMariaRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	//user.Status = 1
	//user.Deleted = 0

	userModel := fromDomain(user)

	err := u.store.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&UserModel{}).
			Where("login_id = ? AND deleted = 0", user.LoginId).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check loginId existence: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("loginId already exists: %s", user.LoginId)
		}

		if err := tx.Model(&UserModel{}).
			Where("email = ? AND deleted = 0", user.Email).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check email existence: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("email already exists: %s", user.Email)
		}

		if err := tx.Create(userModel).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		logger.WithFields(logrus.Fields{
			"user_id":       userModel.Id,
			"login_id":      userModel.LoginId,
			"email":         userModel.Email,
			"rows_affected": tx.RowsAffected,
		}).Info("User created successfully")

		return nil
	})

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"login_id": user.LoginId,
			"email":    user.Email,
		}).Error("Failed to create user in transaction")
		return nil, err
	}

	return userModel.toDomain(), nil
}

func (u *UserMariaRepository) CreateBatch(ctx context.Context, users []*domain.User) error {
	if len(users) == 0 {
		return nil
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	return u.store.Transaction(func(tx *gorm.DB) error {
		userModels := make([]*UserModel, 0, len(users))

		for _, user := range users {
			user.Create()
			user.Status = 1
			user.Deleted = 0
			userModels = append(userModels, fromDomain(user))
		}

		if err := tx.WithContext(ctx).
			CreateInBatches(userModels, 100).Error; err != nil {
			return fmt.Errorf("failed to batch create users: %w", err)
		}

		logger.WithField("count", len(users)).Info("Batch created users successfully")
		return nil
	})
}

func (u *UserMariaRepository) Update(
	ctx context.Context,
	user *domain.User,
	updateFun func(context.Context, *domain.User) (*domain.User, error),
) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.store.Transaction(func(tx *gorm.DB) error {
		var userModel UserModel
		result := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted = 0", user.Id).
			First(&userModel)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return domain.NotFoundError{UserId: user.Id}
			}
			return fmt.Errorf("failed to lock user: %w", result.Error)
		}

		domainUser := userModel.toDomain()

		updatedUser, err := updateFun(ctx, domainUser)
		if err != nil {
			return fmt.Errorf("update function failed: %w", err)
		}

		updatedUser.UpdateTime = time.Now()
		updatedModel := fromDomain(updatedUser)

		if err := tx.WithContext(ctx).
			Model(&UserModel{}).
			Where("id = ?", updatedUser.Id).
			Updates(updatedModel).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		logger.WithFields(logrus.Fields{
			"user_id":       updatedUser.Id,
			"rows_affected": tx.RowsAffected,
		}).Info("User updated successfully")

		return nil
	})
}

func (u *UserMariaRepository) Disable(ctx context.Context, userId int64) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	result := u.store.DB().WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ? AND deleted = 0", userId).
		Updates(map[string]interface{}{
			"status":      0, // setting status to 0: disable
			"update_time": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NotFoundError{UserId: userId}
	}

	logger.WithFields(logrus.Fields{
		"user_id": userId,
		"status":  0,
	}).Info("User status updated")

	return nil
}

func (u *UserMariaRepository) SoftDelete(ctx context.Context, userId int64) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	result := u.store.DB().WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ? AND deleted = 0", userId).
		Updates(map[string]interface{}{
			"deleted":     1, // setting deleted to 1: deleted
			"update_time": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NotFoundError{UserId: userId}
	}

	logger.WithField("user_id", userId).Info("User soft deleted")
	return nil
}

//func (u *UserMariaRepository) Page(ctx context.Context, args pagination.Args) (*pagination.Connection[*domain.User], error) {
//	connection, err := pagination.FetchConnection[UserModel](u.store.DB(), u.store.DB().WithContext(ctx).Model(&UserModel{}), pagination.DefaultConfig(), args)
//	if err != nil {
//		return nil, err
//	}
//	domainConnection := pagination.Connection[*domain.User]{
//		TotalCount: connection.TotalCount,
//		Edges:      toDomainEdges(),
//		PageInfo:   connection.PageInfo,
//	}
//
//}

func (u *UserMariaRepository) ExistsWithTransaction(ctx context.Context, id int64, loginId string, email string) (bool, error) {
	var count int64

	var err error

	if id > 0 {
		err = u.store.DB().Model(&UserModel{}).
			Where("id = ? AND deleted = 0", id).
			Count(&count).Error
	}

	if loginId != "" {
		err = u.store.DB().Model(&UserModel{}).
			Where("login_id = ? AND deleted = 0", loginId).
			Count(&count).Error
	}

	if email != "" {
		err = u.store.DB().Model(&UserModel{}).
			Where("email = ? AND deleted = 0", email).
			Count(&count).Error
	}

	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}

// GetStats query in dashboard or grpc
func (u *UserMariaRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	var stats struct {
		TotalUsers   int64 `gorm:"column:total_users"`
		ActiveUsers  int64 `gorm:"column:active_users"`
		DeletedUsers int64 `gorm:"column:deleted_users"`
	}

	err := u.store.DB().WithContext(ctx).Raw(`
		SELECT 
			COUNT(*) as total_users,
			SUM(CASE WHEN status = 1 AND deleted = 0 THEN 1 ELSE 0 END) as active_users,
			SUM(CASE WHEN status = 0 AND deleted = 0 THEN 1 ELSE 0 END) as disable_users,
			SUM(CASE WHEN deleted = 1 THEN 1 ELSE 0 END) as deleted_users
		FROM users
	`).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return map[string]interface{}{
		"totalUsers":   stats.TotalUsers,
		"activeUsers":  stats.ActiveUsers,
		"disableUsers": stats.ActiveUsers,
		"deletedUsers": stats.DeletedUsers,
	}, nil
}
