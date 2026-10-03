package menu

import (
	"errors"
	"strings"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
)

var ErrInvalidMenu = errors.New("invalid menu")
var ErrMenuNotFound = errors.New("menu not found")
var ErrMenuCodeExist = errors.New("menu code already exists")

// Menu is a catalog resource in the RBAC tree; it is not the permission aggregate.
type Menu struct {
	ID       uint64
	ParentID uint64
	Code     string
	Name     string
	Path     string
	Icon     string
	Sort     int
	Status   int8
}

// Node is a visible menu with the caller's merged bitmask, used to build a tree.
type Node struct {
	ID       uint64   `json:"id"`
	ParentID uint64   `json:"parent_id"`
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Icon     string   `json:"icon"`
	Sort     int      `json:"sort"`
	Status   int8     `json:"status"`
	Mask     uint64   `json:"mask"`
	Actions  []string `json:"actions"`
	Children []*Node  `json:"children"`
}

// New is factory function
// create an active menu catalog item
func New(parentID uint64, code, name, path, icon string, sort int) (*Menu, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, ErrInvalidMenu
	}
	return Reconstruct(0, parentID, code, name, path, icon, sort, 1), nil
}

// Reconstruct hydrates a Menu from persistence without raising events.
func Reconstruct(
	id, parentID uint64,
	code, name, path, icon string,
	sort int,
	status int8,
) *Menu {
	return &Menu{
		ID:       id,
		ParentID: parentID,
		Code:     code,
		Name:     name,
		Path:     path,
		Icon:     icon,
		Sort:     sort,
		Status:   status,
	}
}

// IsActive reports whether the menu is shown in the catalog.
func (m *Menu) IsActive() bool {
	return m != nil && m.Status == 1
}

// ToNode maps a menu plus bitmask into a tree node (READ is decided by BuildTree).
func (m *Menu) ToNode(mask permission.Mask) *Node {
	if m == nil {
		return nil
	}
	return &Node{
		ID:       m.ID,
		ParentID: m.ParentID,
		Code:     m.Code,
		Name:     m.Name,
		Path:     m.Path,
		Icon:     m.Icon,
		Sort:     m.Sort,
		Status:   m.Status,
		Mask:     uint64(mask),
		Actions:  mask.Codes(),
		Children: []*Node{},
	}
}

// BuildTree keeps menus that have READ after mask merge and nests children by parent_id.
func BuildTree(menus []*Menu, matrix map[uint64]permission.Mask) []*Node {
	visible := make(map[uint64]bool, len(menus))
	for _, m := range menus {
		if m == nil || !m.IsActive() {
			continue
		}
		if mask, ok := matrix[m.ID]; ok && mask.Has(permission.ActionRead) {
			visible[m.ID] = true
		}
	}

	nodes := make(map[uint64]*Node, len(menus))
	for _, m := range menus {
		if m == nil || !visible[m.ID] {
			continue
		}
		nodes[m.ID] = m.ToNode(matrix[m.ID])
	}

	var roots []*Node
	for _, m := range menus {
		node, ok := nodes[m.ID]
		if !ok {
			continue
		}
		if m.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodes[m.ParentID]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}
	return roots
}

// BuildCatalog nests every menu, including hidden ones, without permission filtering.
func BuildCatalog(menus []*Menu) []*Node {
	nodes := make(map[uint64]*Node, len(menus))
	for _, m := range menus {
		if m == nil {
			continue
		}
		nodes[m.ID] = m.ToNode(0)
	}
	var roots []*Node
	for _, m := range menus {
		if m == nil {
			continue
		}
		node := nodes[m.ID]
		if m.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodes[m.ParentID]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}
	return roots
}

func (m *Menu) Update(parentID uint64, name, path, icon string, sort int) error {
	if m == nil {
		return ErrInvalidMenu
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidMenu
	}
	m.ParentID = parentID
	m.Name = name
	m.Path = strings.TrimSpace(path)
	m.Icon = strings.TrimSpace(icon)
	m.Sort = sort
	return nil
}

func (m *Menu) ChangeStatus(status int8) error {
	if m == nil {
		return ErrInvalidMenu
	}
	if status != 0 && status != 1 {
		return ErrInvalidMenu
	}
	m.Status = status
	return nil
}
