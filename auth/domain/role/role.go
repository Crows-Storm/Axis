package role

import (
	"errors"
	"time"

	"github.com/Crows-Storm/Axis/auth/domain"
)

// ============================================================
// Role 聚合根
// 职责: 管理角色基本信息 + 角色关联的菜单ID集合
// ============================================================

type RoleCode string

var (
	ADMIN  RoleCode = "ADMIN"
	MANAGE RoleCode = "MANAGE"
	USER   RoleCode = "USER"
)

type Role struct {
	id        int64              // 主键
	roleCode  RoleCode           // 角色编码（值对象）
	roleName  string             // 角色名称
	roleLevel int                // 角色级别，数值越小权限越大
	status    domain.Status      // 状态
	menuIDs   map[int64]struct{} // 已分配的菜单ID集合（聚合内部管理）
	createdAt time.Time
	updatedAt time.Time
}

// ---- 工厂方法 ----

// NewRole 创建新角色（工厂方法，保证创建时即合法）
func NewRole(roleCode RoleCode, roleName string, roleLevel int) (*Role, error) {
	if roleName == "" {
		return nil, errors.New("role name cannot be empty")
	}
	if roleLevel < 0 {
		return nil, errors.New("role level must be >= 0")
	}
	return &Role{
		roleCode:  roleCode,
		roleName:  roleName,
		roleLevel: roleLevel,
		status:    domain.StatusEnabled,
		menuIDs:   make(map[int64]struct{}),
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

// ReconstituteRole 从持久化层重建（仓储调用，不做业务校验）
func ReconstituteRole(id int64, roleCode RoleCode, roleName string,
	roleLevel int, status domain.Status, menuIDs []int64,
	createdAt, updatedAt time.Time) *Role {

	mIDs := make(map[int64]struct{}, len(menuIDs))
	for _, mid := range menuIDs {
		mIDs[mid] = struct{}{}
	}
	return &Role{
		id:        id,
		roleCode:  roleCode,
		roleName:  roleName,
		roleLevel: roleLevel,
		status:    status,
		menuIDs:   mIDs,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ---- 行为方法 ----

// AssignMenus 分配菜单权限（全量替换）
func (r *Role) AssignMenus(menuIDs []int64) {
	r.menuIDs = make(map[int64]struct{}, len(menuIDs))
	for _, mid := range menuIDs {
		r.menuIDs[mid] = struct{}{}
	}
	r.updatedAt = time.Now()
}

// AddMenu 添加单个菜单权限
func (r *Role) AddMenu(menuID int64) {
	r.menuIDs[menuID] = struct{}{}
	r.updatedAt = time.Now()
}

// RemoveMenu 移除单个菜单权限
func (r *Role) RemoveMenu(menuID int64) {
	delete(r.menuIDs, menuID)
	r.updatedAt = time.Now()
}

// HasMenu 判断是否拥有某菜单权限
func (r *Role) HasMenu(menuID int64) bool {
	_, ok := r.menuIDs[menuID]
	return ok
}

// MenuIDs 返回所有菜单ID
func (r *Role) MenuIDs() []int64 {
	ids := make([]int64, 0, len(r.menuIDs))
	for id := range r.menuIDs {
		ids = append(ids, id)
	}
	return ids
}

// Enable / Disable 状态变更
func (r *Role) Enable() {
	r.status = domain.StatusEnabled
	r.updatedAt = time.Now()
}

func (r *Role) Disable() {
	r.status = domain.StatusDisabled
	r.updatedAt = time.Now()
}

// CanOperate 判断当前角色是否有权限操作目标角色
// 规则: 角色级别数值越小权限越大，只能操作级别 >= 自己的角色
func (r *Role) CanOperate(target *Role) bool {
	if !r.status.IsEnabled() {
		return false
	}
	return r.roleLevel <= target.roleLevel
}

// ---- Getters ----

func (r *Role) ID() int64             { return r.id }
func (r *Role) RoleCode() RoleCode    { return r.roleCode }
func (r *Role) RoleName() string      { return r.roleName }
func (r *Role) RoleLevel() int        { return r.roleLevel }
func (r *Role) Status() domain.Status { return r.status }
func (r *Role) CreatedAt() time.Time  { return r.createdAt }
func (r *Role) UpdatedAt() time.Time  { return r.updatedAt }
