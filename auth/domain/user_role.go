package domain

import "time"

// UserRole 用户-角色关联实体
// User 属于 user 限界上下文，这里只持有 userID 引用
type UserRole struct {
	id        int64
	userID    int64
	roleID    int64
	createdAt time.Time
}

func NewUserRole(userID, roleID int64) *UserRole {
	return &UserRole{
		userID:    userID,
		roleID:    roleID,
		createdAt: time.Now(),
	}
}

func ReconstituteUserRole(id, userID, roleID int64, createdAt time.Time) *UserRole {
	return &UserRole{
		id:        id,
		userID:    userID,
		roleID:    roleID,
		createdAt: createdAt,
	}
}

func (ur *UserRole) ID() int64     { return ur.id }
func (ur *UserRole) UserId() int64 { return ur.userID }
func (ur *UserRole) RoleID() int64 { return ur.roleID }
