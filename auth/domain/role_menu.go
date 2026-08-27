package domain

import "time"

// RoleMenu 角色-菜单关联实体
// 作为 Role 聚合内部的关联，也可独立管理
type RoleMenu struct {
	id        int64
	roleID    int64
	menuID    int64
	createdAt time.Time
}

func NewRoleMenu(roleID, menuID int64) *RoleMenu {
	return &RoleMenu{
		roleID:    roleID,
		menuID:    menuID,
		createdAt: time.Now(),
	}
}

func ReconstituteRoleMenu(id, roleID, menuID int64, createdAt time.Time) *RoleMenu {
	return &RoleMenu{
		id:        id,
		roleID:    roleID,
		menuID:    menuID,
		createdAt: createdAt,
	}
}

func (rm *RoleMenu) ID() int64     { return rm.id }
func (rm *RoleMenu) RoleID() int64 { return rm.roleID }
func (rm *RoleMenu) MenuID() int64 { return rm.menuID }
