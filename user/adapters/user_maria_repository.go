package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	"github.com/Crows-Storm/Axis/common/server/store"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModel struct {
	ID       int64  `gorm:"column:id;unsigned;primaryKey;"`
	LoginId  string `gorm:"column:login_id;uniqueIndex:uk_login_id;type:varchar(50);not null"`
	Username string `gorm:"column:username;type:varchar(50);not null"`
	Password string `gorm:"column:password;type:varchar(255);not null"`
	Email    string `gorm:"column:email;uniqueIndex:uk_email;type:varchar(100);not null"`
	// status: 0: pending 1: activated 2: disabled
	Status     int8      `gorm:"column:status;type:tinyint;default:1;not null;index:idx_status;index:idx_deleted_status,priority:2"`
	Deleted    int8      `gorm:"column:deleted;type:tinyint;default:0;not null;index:idx_deleted_status,priority:1"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null;autoUpdateTime"`
}

func (*UserModel) TableName() string {
	return "sys_user"
}

func (m *UserModel) toDomain() *domain.User {
	// TODO: maybe need apply event to build Domain: Event sourcing
	return &domain.User{
		ID:      m.ID,
		LoginId: m.LoginId,
		Email:   m.Email,
		Status:  commuser.Status(m.Status),
	}
}

func fromDomain(user *domain.User) *UserModel {
	return &UserModel{
		ID:      user.ID,
		LoginId: user.LoginId,
		Email:   user.Email,
		Status:  user.Status.Value(),
	}
}

type UserMariaRepository struct {
	store *store.Store
}

func NewUserMariaRepository(store *store.Store) *UserMariaRepository {
	repo := &UserMariaRepository{
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
	var userModel UserModel
	result := u.store.DB().Scopes(domain.NotDeleted).
		Where("id = ?", id).
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
	var userModel UserModel
	result := u.store.DB().WithContext(ctx).Scopes(domain.NotDeleted).
		Where("login_id = ?", loginId).
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
	userModel := fromDomain(user)

	err := u.store.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&UserModel{}).Scopes(domain.NotDeleted).
			Where("login_id = ?", user.LoginId).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check loginId existence: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("loginId already exists: %s", user.LoginId)
		}

		if err := tx.Model(&UserModel{}).Scopes(domain.NotDeleted).
			Where("email = ?", user.Email).
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
			"user_id":       userModel.ID,
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

	return u.store.Transaction(func(tx *gorm.DB) error {
		userModels := make([]*UserModel, 0, len(users))

		for _, user := range users {
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

func (u *UserMariaRepository) Update(ctx context.Context, user *domain.User) error {
	return u.store.Transaction(func(tx *gorm.DB) error {
		var userModel UserModel
		result := tx.WithContext(ctx).Scopes(domain.NotDeleted).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", user.ID).
			First(&userModel)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return domain.NotFoundError{UserId: user.ID}
			}
			return fmt.Errorf("failed to lock user: %w", result.Error)
		}

		updatedModel := fromDomain(user)

		if err := tx.WithContext(ctx).
			Model(&UserModel{}).
			Where("id = ?", user.ID).
			Updates(updatedModel).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		logger.WithFields(logrus.Fields{
			"user_id":       user.ID,
			"rows_affected": tx.RowsAffected,
		}).Info("User updated successfully")

		if tx.RowsAffected < 1 {
			return domain.NotFoundError{UserId: user.ID}
		}
		return nil
	})
}

func (u *UserMariaRepository) Disable(ctx context.Context, userId int64) error {
	result := u.store.DB().WithContext(ctx).
		Model(&UserModel{}).Scopes(domain.NotDeleted).
		Where("id = ?", userId).
		Updates(map[string]interface{}{
			"status":      commuser.Disabled, // setting status to 2: disable
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
		"status":  commuser.Disabled,
	}).Info("User status updated")

	return nil
}

func (u *UserMariaRepository) SoftDelete(ctx context.Context, userId int64) error {
	result := u.store.DB().WithContext(ctx).
		Model(&UserModel{}).Scopes(domain.NotDeleted).
		Where("id = ?", userId).
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

func (u *UserMariaRepository) GetPasswordByLoginId(ctx context.Context, loginId string) string {
	user, err := u.GetByLoginId(ctx, loginId)
	if err != nil {
		return ""
	}
	return user.Password
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
		err = u.store.DB().Model(&UserModel{}).Scopes(domain.NotDeleted).
			Where("id = ?", id).
			Count(&count).Error
	}

	if loginId != "" {
		err = u.store.DB().Model(&UserModel{}).Scopes(domain.NotDeleted).
			Where("login_id = ?", loginId).
			Count(&count).Error
	}

	if email != "" {
		err = u.store.DB().Model(&UserModel{}).Scopes(domain.NotDeleted).
			Where("email = ?", email).
			Count(&count).Error
	}

	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}

// GetStats query in dashboard or grpc
func (u *UserMariaRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		TotalUsers   int64 `gorm:"column:total_users"`
		ActiveUsers  int64 `gorm:"column:active_users"`
		DeletedUsers int64 `gorm:"column:deleted_users"`
	}

	err := u.store.DB().WithContext(ctx).Raw(`
		SELECT 
			COUNT(*) as total_users,
			SUM(CASE WHEN status = 1 AND deleted = 0 THEN 1 ELSE 0 END) as active_users,
			SUM(CASE WHEN status = 2 AND deleted = 0 THEN 1 ELSE 0 END) as disable_users,
			SUM(CASE WHEN deleted = 1 THEN 1 ELSE 0 END) as deleted_users
		FROM sys_user
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
