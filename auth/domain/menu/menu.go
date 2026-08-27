package menu

import (
	"errors"
	"time"

	"github.com/Crows-Storm/Axis/auth/domain"
)

// ============================================================
// Menu 聚合根
// 职责: 管理菜单/权限节点，维护树形结构
// ============================================================

type Menu struct {
	id         int64         // 主键
	parentID   int64         // 父菜单ID, 0为顶级
	menuCode   string        // 菜单编码
	menuName   string        // 菜单名称
	menuType   MenuType      // 菜单类型
	path       string        // 路由路径
	treePath   TreePath      // 树路径
	component  string        // 组件路径
	icon       string        // 图标
	permission Permission    // 权限标识
	sortOrder  int           // 排序号
	visible    bool          // 是否显示
	status     domain.Status // 状态
	createdAt  time.Time
	updatedAt  time.Time
}

// ---- 工厂方法 ----

func NewMenu(menuCode, menuName string, menuType MenuType, parentID int64) (*Menu, error) {
	if menuCode == "" {
		return nil, errors.New("menu code cannot be empty")
	}
	if menuName == "" {
		return nil, errors.New("menu name cannot be empty")
	}
	return &Menu{
		parentID:  parentID,
		menuCode:  menuCode,
		menuName:  menuName,
		menuType:  menuType,
		status:    domain.StatusEnabled,
		visible:   true,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

// ReconstituteMenu 从持久化层重建
func ReconstituteMenu(id, parentID int64, menuCode, menuName string,
	menuType MenuType, path string, treePath TreePath,
	component, icon string, permission Permission,
	sortOrder int, visible bool, status domain.Status,
	createdAt, updatedAt time.Time) *Menu {

	return &Menu{
		id:         id,
		parentID:   parentID,
		menuCode:   menuCode,
		menuName:   menuName,
		menuType:   menuType,
		path:       path,
		treePath:   treePath,
		component:  component,
		icon:       icon,
		permission: permission,
		sortOrder:  sortOrder,
		visible:    visible,
		status:     status,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// ---- 行为方法 ----

// SetRouteInfo 设置路由信息（仅目录和菜单类型可设置）
func (m *Menu) SetRouteInfo(path, component, icon string) error {
	if m.menuType.IsButton() {
		return errors.New("button type menu cannot have route info")
	}
	m.path = path
	m.component = component
	m.icon = icon
	m.updatedAt = time.Now()
	return nil
}

// SetPermission 设置权限标识（仅按钮类型必须设置）
func (m *Menu) SetPermission(perm Permission) error {
	if perm.IsEmpty() && m.menuType.IsButton() {
		return errors.New("button type menu must have permission identifier")
	}
	m.permission = perm
	m.updatedAt = time.Now()
	return nil
}

// UpdateTreePath 更新树路径（在持久化后由领域服务调用）
func (m *Menu) UpdateTreePath(parentPath TreePath) {
	m.treePath = BuildTreePath(parentPath, m.id)
	m.updatedAt = time.Now()
}

func (m *Menu) ChangeSortOrder(order int) {
	m.sortOrder = order
	m.updatedAt = time.Now()
}

func (m *Menu) Show()    { m.visible = true; m.updatedAt = time.Now() }
func (m *Menu) Hide()    { m.visible = false; m.updatedAt = time.Now() }
func (m *Menu) Enable()  { m.status = domain.StatusEnabled; m.updatedAt = time.Now() }
func (m *Menu) Disable() { m.status = domain.StatusDisabled; m.updatedAt = time.Now() }

// IsTopLevel 是否顶级菜单
func (m *Menu) IsTopLevel() bool {
	return m.parentID == 0
}

func (m *Menu) ID() int64              { return m.id }
func (m *Menu) ParentID() int64        { return m.parentID }
func (m *Menu) MenuCode() string       { return m.menuCode }
func (m *Menu) MenuName() string       { return m.menuName }
func (m *Menu) MenuType() MenuType     { return m.menuType }
func (m *Menu) Path() string           { return m.path }
func (m *Menu) TreePath() TreePath     { return m.treePath }
func (m *Menu) Component() string      { return m.component }
func (m *Menu) Icon() string           { return m.icon }
func (m *Menu) Permission() Permission { return m.permission }
func (m *Menu) SortOrder() int         { return m.sortOrder }
func (m *Menu) IsVisible() bool        { return m.visible }
func (m *Menu) Status() domain.Status  { return m.status }
func (m *Menu) CreatedAt() time.Time   { return m.createdAt }
func (m *Menu) UpdatedAt() time.Time   { return m.updatedAt }
