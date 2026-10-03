package role

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/common/domain"
	"github.com/Crows-Storm/Axis/common/domain/event"
	commrole "github.com/Crows-Storm/Axis/common/domain/role"
	"github.com/Crows-Storm/Axis/common/domain/role/events"
	"github.com/Crows-Storm/Axis/common/util"
)

// MenuGrant is a role's permission bitmask on one menu resource.
type MenuGrant struct {
	MenuID uint64
	Mask   permission.Mask
}

// Role is the RBAC aggregate root: identity plus menu grants.
type Role struct {
	domain.AggregateRoot

	ID          uint64          `json:"id"`
	Code        string          `json:"code"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      commrole.Status `json:"status"`
	grants      map[uint64]permission.Mask
}

var (
	_ domain.Aggregate = (*Role)(nil)
	_ domain.Applier   = (*Role)(nil)
)

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrInvalidRole       = errors.New("invalid role")
	ErrRoleDisabled      = errors.New("role is disabled")
	ErrInvalidAssignment = errors.New("invalid user-role assignment")
	ErrInvalidGrant      = errors.New("invalid menu grant")
	ErrUserRoleExist     = errors.New("user already has this role")
	ErrRoleCodeExist     = errors.New("role code already exists")
	ErrGrantNotFound     = errors.New("role menu grant not found")
)

// The Create is factory function
// create role and raise RoleCreated domain event
func Create(code, name, description string) (*Role, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, ErrInvalidRole
	}

	r := &Role{grants: make(map[uint64]permission.Mask)}
	if err := r.RaiseEvent(r, events.NewRoleCreatedMessage(
		util.GenerateFlakeID(),
		code,
		name,
		description,
	)); err != nil { // setting event to root.events and increment version
		return nil, fmt.Errorf("role.Create: failed to apply domain event, %w", err)
	}
	return r, nil
}

// Reconstruct hydrates a Role from persistence without raising events.
func Reconstruct(
	id uint64,
	code, name, description string,
	status commrole.Status,
	grants map[uint64]permission.Mask,
) *Role {
	if grants == nil {
		grants = make(map[uint64]permission.Mask)
	}
	copied := make(map[uint64]permission.Mask, len(grants))
	for menuID, mask := range grants {
		copied[menuID] = mask
	}
	return &Role{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: description,
		Status:      status,
		grants:      copied,
	}
}

// IsActive reports whether the role can be assigned or granted menus.
func (r *Role) IsActive() bool {
	return r != nil && r.Status.IsActive()
}

// Grants returns a copy of menu permission grants owned by this aggregate.
func (r *Role) Grants() []MenuGrant {
	if r == nil {
		return nil
	}
	out := make([]MenuGrant, 0, len(r.grants))
	for menuID, mask := range r.grants {
		out = append(out, MenuGrant{MenuID: menuID, Mask: mask})
	}
	return out
}

// MaskOf returns the permission bitmask for a menu; missing grants are 0.
func (r *Role) MaskOf(menuID uint64) permission.Mask {
	if r == nil {
		return 0
	}
	return r.grants[menuID]
}

// AssignTo binds a user to this role and raises UserAssignedToRole.
func (r *Role) AssignTo(userID uint64) error {
	if r == nil {
		return ErrInvalidRole
	}
	if userID == 0 {
		return ErrInvalidAssignment
	}
	if !r.IsActive() {
		return ErrRoleDisabled
	}
	if err := r.RaiseEvent(r, events.NewUserAssignedToRoleMessage(r.ID, userID)); err != nil { // setting event to root.events and increment version
		return fmt.Errorf("role.AssignTo: failed to apply domain event, %w", err)
	}
	return nil
}

// GrantMenu OR-merges bits onto an existing menu grant and raises RoleMenuGranted.
func (r *Role) GrantMenu(menuID uint64, mask permission.Mask) error {
	if err := r.ensureWritableGrant(menuID); err != nil {
		return err
	}
	if err := r.RaiseEvent(r, events.NewRoleMenuGrantedMessage(r.ID, menuID, uint64(mask))); err != nil { // setting event to root.events and increment version
		return fmt.Errorf("role.GrantMenu: failed to apply domain event, %w", err)
	}
	return nil
}

// ReplaceMenuMask overwrites the menu grant mask and raises RoleMenuMaskReplaced.
func (r *Role) ReplaceMenuMask(menuID uint64, mask permission.Mask) error {
	if err := r.ensureWritableGrant(menuID); err != nil {
		return err
	}
	if err := r.RaiseEvent(r, events.NewRoleMenuMaskReplacedMessage(r.ID, menuID, uint64(mask))); err != nil { // setting event to root.events and increment version
		return fmt.Errorf("role.ReplaceMenuMask: failed to apply domain event, %w", err)
	}
	return nil
}

// ensureWritableGrant checks invariants before mutating menu grants.
func (r *Role) ensureWritableGrant(menuID uint64) error {
	if r == nil {
		return ErrInvalidRole
	}
	if menuID == 0 {
		return ErrInvalidGrant
	}
	if !r.IsActive() {
		return ErrRoleDisabled
	}
	if r.grants == nil {
		r.grants = make(map[uint64]permission.Mask)
	}
	return nil
}

// UpdateProfile changes display fields and raises RoleUpdated.
func (r *Role) UpdateProfile(name, description string) error {
	if r == nil {
		return ErrInvalidRole
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidRole
	}
	if err := r.RaiseEvent(r, events.NewRoleUpdatedMessage(r.ID, name, strings.TrimSpace(description))); err != nil {
		return fmt.Errorf("role.UpdateProfile: failed to apply domain event, %w", err)
	}
	return nil
}

// ChangeStatus enables or disables the role.
func (r *Role) ChangeStatus(status commrole.Status) error {
	if r == nil {
		return ErrInvalidRole
	}
	if status != commrole.Active && status != commrole.Disabled {
		return ErrInvalidRole
	}
	if err := r.RaiseEvent(r, events.NewRoleStatusChangedMessage(r.ID, int8(status))); err != nil {
		return fmt.Errorf("role.ChangeStatus: failed to apply domain event, %w", err)
	}
	return nil
}

// RevokeMenu removes a menu grant.
func (r *Role) RevokeMenu(menuID uint64) error {
	if err := r.ensureWritableGrant(menuID); err != nil {
		return err
	}
	if _, ok := r.grants[menuID]; !ok {
		return ErrGrantNotFound
	}
	if err := r.RaiseEvent(r, events.NewRoleMenuRevokedMessage(r.ID, menuID)); err != nil {
		return fmt.Errorf("role.RevokeMenu: failed to apply domain event, %w", err)
	}
	return nil
}

// UnassignFrom unbinds a user from this role.
func (r *Role) UnassignFrom(userID uint64) error {
	if r == nil {
		return ErrInvalidRole
	}
	if userID == 0 {
		return ErrInvalidAssignment
	}
	if err := r.RaiseEvent(r, events.NewUserUnassignedFromRoleMessage(r.ID, userID)); err != nil {
		return fmt.Errorf("role.UnassignFrom: failed to apply domain event, %w", err)
	}
	return nil
}

// ApplyEvent projects a domain event onto Role state (Applier).
func (r *Role) ApplyEvent(evt event.Message) error {
	if r.grants == nil {
		r.grants = make(map[uint64]permission.Mask)
	}

	switch e := evt.(type) {
	case *events.RoleCreatedMessage:
		r.ID = e.AggregateID
		r.Code = e.Code
		r.Name = e.RoleName
		r.Description = e.Description
		r.Status = commrole.Active
	case *events.RoleMenuGrantedMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
		r.grants[e.MenuID] |= permission.Mask(e.Mask)
	case *events.RoleMenuMaskReplacedMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
		r.grants[e.MenuID] = permission.Mask(e.Mask)
	case *events.UserAssignedToRoleMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
	case *events.RoleUpdatedMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
		r.Name = e.RoleName
		r.Description = e.Description
	case *events.RoleStatusChangedMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
		r.Status = commrole.Status(e.Status)
	case *events.RoleMenuRevokedMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
		delete(r.grants, e.MenuID)
	case *events.UserUnassignedFromRoleMessage:
		if e.AggregateID != 0 {
			r.ID = e.AggregateID
		}
	}
	return nil
}

// MergeGrants OR-merges menu masks across active roles (multi-role permission union).
func MergeGrants(roles []*Role) map[uint64]permission.Mask {
	matrix := make(map[uint64]permission.Mask)
	for _, r := range roles {
		if r == nil || !r.IsActive() {
			continue
		}
		for menuID, mask := range r.grants {
			matrix[menuID] |= mask
		}
	}
	return matrix
}
