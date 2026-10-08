package menu

import (
	"testing"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/stretchr/testify/assert"
)

func TestBuildTree_OnlyReadableMenus(t *testing.T) {
	menus := []*Menu{
		Reconstruct(1, 0, "user", "user", "", "", 1, 1),
		Reconstruct(2, 0, "system", "system", "", "", 2, 1),
		Reconstruct(10, 1, "user-detail", "detail", "", "", 1, 1),
	}
	matrix := map[uint64]permission.Mask{
		1:  permission.PermRead,
		10: permission.PermCRUD,
	}
	tree := BuildTree(menus, matrix)
	assert.Len(t, tree, 1)
	assert.Equal(t, "user", tree[0].Code)
	assert.Len(t, tree[0].Children, 1)
	assert.Contains(t, tree[0].Children[0].Actions, "CREATE")
}

func TestBuildCatalog_IncludesHidden(t *testing.T) {
	menus := []*Menu{
		Reconstruct(1, 0, "user", "user", "", "", 1, 1),
		Reconstruct(2, 0, "hidden", "hidden", "", "", 2, 0),
	}
	tree := BuildCatalog(menus)
	assert.Len(t, tree, 2)
}
