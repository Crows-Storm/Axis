package rbac

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type Authorizer struct {
	store  Backend
	source Source
}

func NewAuthorizer(store Backend, source Source) *Authorizer {
	if store == nil {
		return nil
	}
	return &Authorizer{store: store, source: source}
}

type Request struct {
	UserID   uint64
	Audience Audience
	Method   string
	Path     string
}

func (a *Authorizer) Authorize(ctx context.Context, req Request) error {
	if a == nil || a.store == nil {
		return ErrUnavailable
	}
	if req.UserID == 0 {
		return ErrUnauthenticated
	}
	audience := NormalizeAudience(string(req.Audience))
	appReq, err := a.lookupRoute(ctx, AudienceApp, req.Method, req.Path)
	if err != nil {
		return err
	}
	consoleReq, err := a.lookupRoute(ctx, AudienceConsole, req.Method, req.Path)
	if err != nil {
		return err
	}
	switch audience {
	case AudienceApp:
		if appReq != nil {
			return a.enforce(ctx, req.UserID, *appReq)
		}
		if consoleReq != nil {
			return ErrForbidden
		}
		return nil
	case AudienceConsole:
		if consoleReq != nil {
			return a.enforce(ctx, req.UserID, *consoleReq)
		}
		if appReq != nil && appReq.Self {
			return nil
		}
		return ErrForbidden
	default:
		return ErrForbidden
	}
}

func (a *Authorizer) UserMenuMask(ctx context.Context, userID uint64, menuCode string) (Mask, error) {
	if a == nil || a.store == nil {
		return 0, ErrUnavailable
	}
	roles, err := a.userRoles(ctx, userID)
	if err != nil {
		return 0, err
	}
	var merged Mask
	for _, roleID := range roles {
		part, err := a.roleMenuMask(ctx, roleID, menuCode)
		if err != nil {
			return 0, err
		}
		merged |= part
	}
	return merged, nil
}

func (a *Authorizer) InvalidateUser(ctx context.Context, userID uint64) error {
	if a == nil || a.store == nil || userID == 0 {
		return nil
	}
	if err := a.store.Del(ctx, userRolesKey(userID)); err != nil {
		return fmt.Errorf("invalidate user roles: %w", err)
	}
	return nil
}

func (a *Authorizer) InvalidateRole(ctx context.Context, roleID uint64) error {
	if a == nil || a.store == nil || roleID == 0 {
		return nil
	}
	if err := a.store.Del(ctx, roleKey(roleID)); err != nil {
		return fmt.Errorf("invalidate role: %w", err)
	}
	return nil
}

func (a *Authorizer) InvalidateMenu(ctx context.Context, menuCode string) error {
	if a == nil || a.store == nil || strings.TrimSpace(menuCode) == "" {
		return nil
	}
	if a.source == nil {
		return nil
	}
	ids, err := a.source.RoleIDsForMenuCode(ctx, menuCode)
	if err != nil {
		return fmt.Errorf("roles for menu: %w", err)
	}
	for _, id := range ids {
		if err := a.InvalidateRole(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (a *Authorizer) WarmUser(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return nil
	}
	ids, err := a.userRoles(ctx, userID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := a.roleMenuMask(ctx, id, roleLoadedField); err != nil {
			return err
		}
	}
	return nil
}

func (a *Authorizer) PublishRoutes(ctx context.Context, routes []Route) error {
	if a == nil || a.store == nil {
		return ErrUnavailable
	}
	for _, route := range routes {
		raw, err := route.Req.encode()
		if err != nil {
			return err
		}
		if err := a.store.Set(ctx, routeKey(route.Audience, route.Method, route.Path), raw); err != nil {
			return err
		}
	}
	return nil
}

func (a *Authorizer) lookupRoute(ctx context.Context, audience Audience, method, path string) (*Requirement, error) {
	raw, ok, err := a.store.Get(ctx, routeKey(audience, method, path))
	if err != nil {
		return nil, ErrUnavailable
	}
	if !ok {
		return nil, nil
	}
	req, err := decodeRequirement(raw)
	if err != nil {
		return nil, ErrUnavailable
	}
	return &req, nil
}

func (a *Authorizer) enforce(ctx context.Context, userID uint64, req Requirement) error {
	if req.Self {
		return nil
	}
	action, ok := ParseAction(req.Action)
	if !ok {
		return ErrUnavailable
	}
	mask, err := a.UserMenuMask(ctx, userID, req.Menu)
	if err != nil {
		return err
	}
	if !mask.Has(action) {
		return ErrForbidden
	}
	return nil
}

func (a *Authorizer) userRoles(ctx context.Context, userID uint64) ([]uint64, error) {
	key := userRolesKey(userID)
	exists, err := a.store.Exists(ctx, key)
	if err != nil {
		return nil, ErrUnavailable
	}
	if !exists {
		if a.source == nil {
			return nil, nil
		}
		ids, err := a.source.EnabledRoleIDs(ctx, userID)
		if err != nil {
			return nil, ErrUnavailable
		}
		if err := a.writeUserRoles(ctx, userID, ids); err != nil {
			return nil, ErrUnavailable
		}
		return ids, nil
	}
	members, err := a.store.SMembers(ctx, key)
	if err != nil {
		return nil, ErrUnavailable
	}
	out := make([]uint64, 0, len(members))
	for _, member := range members {
		if member == emptyRoleMember {
			continue
		}
		id, err := strconv.ParseUint(member, 10, 64)
		if err != nil || id == 0 {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

func (a *Authorizer) writeUserRoles(ctx context.Context, userID uint64, ids []uint64) error {
	key := userRolesKey(userID)
	_ = a.store.Del(ctx, key)
	if len(ids) == 0 {
		return a.store.SAdd(ctx, key, emptyRoleMember)
	}
	members := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		members = append(members, strconv.FormatUint(id, 10))
	}
	if len(members) == 0 {
		return a.store.SAdd(ctx, key, emptyRoleMember)
	}
	return a.store.SAdd(ctx, key, members...)
}

func (a *Authorizer) roleMenuMask(ctx context.Context, roleID uint64, menuCode string) (Mask, error) {
	key := roleKey(roleID)
	exists, err := a.store.Exists(ctx, key)
	if err != nil {
		return 0, ErrUnavailable
	}
	if !exists {
		if a.source == nil {
			return 0, nil
		}
		grants, err := a.source.RoleMenuMasks(ctx, roleID)
		if err != nil {
			return 0, ErrUnavailable
		}
		fields := map[string]string{roleLoadedField: "1"}
		for code, mask := range grants {
			fields[code] = strconv.FormatUint(mask, 10)
		}
		if err := a.store.HSet(ctx, key, fields); err != nil {
			return 0, ErrUnavailable
		}
	}
	raw, ok, err := a.store.HGet(ctx, key, menuCode)
	if err != nil {
		return 0, ErrUnavailable
	}
	if !ok || raw == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, ErrUnavailable
	}
	return Mask(n), nil
}
