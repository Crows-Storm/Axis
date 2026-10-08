package role

import (
	"testing"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	commrole "github.com/Crows-Storm/Axis/common/domain/role"
	"github.com/Crows-Storm/Axis/common/domain/role/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRole_GrantAndAssignRaiseEvents(t *testing.T) {
	admin, err := Create("admin", "超级管理员", "")
	require.NoError(t, err)
	require.True(t, admin.HasEvents())
	require.Equal(t, "RoleCreated", admin.Events()[0].Message.Name())

	loaded := Reconstruct(admin.ID, admin.Code, admin.Name, admin.Description, commrole.Active, nil)
	assert.False(t, loaded.HasEvents())

	require.NoError(t, loaded.GrantMenu(3, permission.PermRead|permission.PermCreate|permission.PermUpdate|permission.PermEnable|permission.PermClose))
	assert.Equal(t, permission.Mask(55), loaded.MaskOf(3))
	assert.False(t, loaded.MaskOf(3).Has(permission.ActionDelete))
	require.Len(t, loaded.Events(), 1)
	_, ok := loaded.Events()[0].Message.(*events.RoleMenuGrantedMessage)
	assert.True(t, ok)

	require.NoError(t, loaded.GrantMenu(3, permission.PermDelete))
	assert.True(t, loaded.MaskOf(3).Has(permission.ActionDelete))
	require.Len(t, loaded.Events(), 2)

	require.NoError(t, loaded.ReplaceMenuMask(1, permission.PermReadOnly))
	assert.Equal(t, permission.PermReadOnly, loaded.MaskOf(1))
	require.Equal(t, "RoleMenuMaskReplaced", loaded.Events()[2].Message.Name())

	require.NoError(t, loaded.AssignTo(2))
	require.Equal(t, "UserAssignedToRole", loaded.Events()[3].Message.Name())

	assert.ErrorIs(t, loaded.AssignTo(0), ErrInvalidAssignment)
	assert.ErrorIs(t, loaded.GrantMenu(0, permission.PermRead), ErrInvalidGrant)
	assert.Len(t, loaded.Events(), 4)

	disabled := Reconstruct(9, "guest", "访客", "", commrole.Disabled, nil)
	assert.ErrorIs(t, disabled.AssignTo(1), ErrRoleDisabled)
	assert.ErrorIs(t, disabled.GrantMenu(1, permission.PermRead), ErrRoleDisabled)
	assert.False(t, disabled.HasEvents())

	require.NoError(t, loaded.RevokeMenu(1))
	assert.Equal(t, permission.Mask(0), loaded.MaskOf(1))
	require.Equal(t, "RoleMenuRevoked", loaded.Events()[len(loaded.Events())-1].Message.Name())
	assert.ErrorIs(t, loaded.RevokeMenu(1), ErrGrantNotFound)

	require.NoError(t, loaded.UpdateProfile("编辑改", "desc"))
	assert.Equal(t, "编辑改", loaded.Name)
	require.NoError(t, loaded.ChangeStatus(commrole.Disabled))
	assert.False(t, loaded.IsActive())
	require.NoError(t, loaded.UnassignFrom(2))
}

func TestMergeGrants_ORAcrossRoles(t *testing.T) {
	editor := Reconstruct(2, "editor", "编辑", "", commrole.Active, map[uint64]permission.Mask{
		1: permission.PermRead,
		3: permission.PermRead | permission.PermCreate | permission.PermUpdate,
	})
	auditor := Reconstruct(3, "auditor", "审核", "", commrole.Active, map[uint64]permission.Mask{
		3: permission.PermAudit,
	})
	off := Reconstruct(4, "off", "停用", "", commrole.Disabled, map[uint64]permission.Mask{
		3: permission.PermAll,
	})

	matrix := MergeGrants([]*Role{editor, auditor, off})
	assert.Equal(t, permission.PermRead, matrix[1])
	assert.True(t, matrix[3].Has(permission.ActionAudit))
	assert.True(t, matrix[3].Has(permission.ActionUpdate))
	assert.False(t, matrix[3].Has(permission.ActionDelete))
}
