package rbac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubSource struct {
	roles  map[uint64][]uint64
	grants map[uint64]map[string]uint64
}

func (s stubSource) EnabledRoleIDs(_ context.Context, userID uint64) ([]uint64, error) {
	return s.roles[userID], nil
}

func (s stubSource) RoleMenuMasks(_ context.Context, roleID uint64) (map[string]uint64, error) {
	return s.grants[roleID], nil
}

func (s stubSource) RoleIDsForMenuCode(_ context.Context, _ string) ([]uint64, error) {
	return nil, nil
}

func TestAuthorizeAppSelfAndConsoleCheck(t *testing.T) {
	store := NewMemory()
	src := stubSource{
		roles: map[uint64][]uint64{9: {3}},
		grants: map[uint64]map[string]uint64{
			3: {"account": 1}, // READ
		},
	}
	authz := NewAuthorizer(store, src)
	require.NoError(t, authz.PublishRoutes(context.Background(), []Route{
		Self(AudienceApp, "GET", "/api/account"),
		Need(AudienceConsole, "GET", "/api/account/:id", "account", "READ"),
	}))

	require.NoError(t, authz.Authorize(context.Background(), Request{
		UserID: 9, Audience: AudienceApp, Method: "GET", Path: "/api/account",
	}))
	require.ErrorIs(t, authz.Authorize(context.Background(), Request{
		UserID: 9, Audience: AudienceApp, Method: "GET", Path: "/api/account/:id",
	}), ErrForbidden)
	require.NoError(t, authz.Authorize(context.Background(), Request{
		UserID: 9, Audience: AudienceConsole, Method: "GET", Path: "/api/account/:id",
	}))
	require.ErrorIs(t, authz.Authorize(context.Background(), Request{
		UserID: 8, Audience: AudienceConsole, Method: "GET", Path: "/api/account/:id",
	}), ErrForbidden)
	require.NoError(t, authz.Authorize(context.Background(), Request{
		UserID: 9, Audience: AudienceConsole, Method: "GET", Path: "/api/account",
	}))
}

func TestUserMenuMaskMergesRolesFromCache(t *testing.T) {
	store := NewMemory()
	src := stubSource{
		roles: map[uint64][]uint64{1: {10, 11}},
		grants: map[uint64]map[string]uint64{
			10: {"user": 1},
			11: {"user": 2},
		},
	}
	authz := NewAuthorizer(store, src)
	mask, err := authz.UserMenuMask(context.Background(), 1, "user")
	require.NoError(t, err)
	require.True(t, mask.Has(ActionRead))
	require.True(t, mask.Has(ActionCreate))

	src.grants[11] = map[string]uint64{"user": 0}
	mask, err = authz.UserMenuMask(context.Background(), 1, "user")
	require.NoError(t, err)
	require.True(t, mask.Has(ActionCreate), "cached role hash must be used until invalidate")
}
