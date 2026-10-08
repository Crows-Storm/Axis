package adapters

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	"github.com/Crows-Storm/Axis/common/server/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionMatrix_MultiRoleOR(t *testing.T) {
	st, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: filepath.Join(t.TempDir(), "rbac.db"),
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	roleRepo := NewRoleMariaRepository(st)
	menuRepo := NewMenuMariaRepository(st)
	permRepo := NewPermissionMariaRepository(st)
	ctx := context.Background()

	require.NoError(t, st.DB().Create(&RoleModel{ID: 1, Code: "admin", Name: "admin", Status: 1}).Error)
	require.NoError(t, st.DB().Create(&RoleModel{ID: 2, Code: "editor", Name: "editor", Status: 1}).Error)
	require.NoError(t, st.DB().Create(&MenuModel{ID: 1, Code: "user", Name: "user", Status: 1, Sort: 1}).Error)
	require.NoError(t, st.DB().Create(&MenuModel{ID: 3, Code: "product", Name: "product", Status: 1, Sort: 2}).Error)

	require.NoError(t, replaceMask(ctx, roleRepo, 1, 1, permission.PermAll))
	require.NoError(t, replaceMask(ctx, roleRepo, 1, 3, permission.PermAll))
	require.NoError(t, replaceMask(ctx, roleRepo, 2, 1, permission.PermRead))
	require.NoError(t, replaceMask(ctx, roleRepo, 2, 3, permission.PermRead|permission.PermCreate|permission.PermUpdate|permission.PermEnable|permission.PermClose))

	admin, err := roleRepo.Get(ctx, 1)
	require.NoError(t, err)
	require.NoError(t, roleRepo.AssignUser(ctx, 1, admin))
	editor, err := roleRepo.Get(ctx, 2)
	require.NoError(t, err)
	require.NoError(t, roleRepo.AssignUser(ctx, 2, editor))

	aliceUser, err := permRepo.GetUserMenuMask(ctx, 1, "user")
	require.NoError(t, err)
	assert.True(t, aliceUser.Has(permission.ActionDelete))

	bobProduct, err := permRepo.GetUserMenuMask(ctx, 2, "product")
	require.NoError(t, err)
	assert.True(t, bobProduct.Has(permission.ActionUpdate))
	assert.False(t, bobProduct.Has(permission.ActionDelete))

	bobSystem, err := permRepo.GetUserMenuMask(ctx, 2, "system")
	require.NoError(t, err)
	assert.False(t, bobSystem.Has(permission.ActionRead))

	require.NoError(t, grantMask(ctx, roleRepo, 2, 3, permission.PermDelete))
	bobProduct, err = permRepo.GetUserMenuMask(ctx, 2, "product")
	require.NoError(t, err)
	assert.True(t, bobProduct.Has(permission.ActionDelete))

	loaded, err := roleRepo.Get(ctx, 2)
	require.NoError(t, err)
	assert.True(t, loaded.MaskOf(3).Has(permission.ActionDelete))

	menus, err := menuRepo.ListActive(ctx)
	require.NoError(t, err)
	matrix, err := permRepo.GetUserPermissionMatrix(ctx, 1)
	require.NoError(t, err)
	tree := menu.BuildTree(menus, matrix)
	require.Len(t, tree, 2)
}

func replaceMask(ctx context.Context, repo *RoleMariaRepository, roleID, menuID uint64, mask permission.Mask) error {
	return repo.Update(ctx, &role.Role{ID: roleID}, func(ctx context.Context, current *role.Role) (*role.Role, error) {
		if err := current.ReplaceMenuMask(menuID, mask); err != nil {
			return nil, err
		}
		return current, nil
	})
}

func grantMask(ctx context.Context, repo *RoleMariaRepository, roleID, menuID uint64, mask permission.Mask) error {
	return repo.Update(ctx, &role.Role{ID: roleID}, func(ctx context.Context, current *role.Role) (*role.Role, error) {
		if err := current.GrantMenu(menuID, mask); err != nil {
			return nil, err
		}
		return current, nil
	})
}
