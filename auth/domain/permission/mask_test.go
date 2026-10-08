package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMask_HasAndCombine(t *testing.T) {
	crud := PermCRUD
	assert.True(t, crud.Has(ActionRead))
	assert.True(t, crud.Has(ActionDelete))
	assert.False(t, crud.Has(ActionExport))
	assert.True(t, crud.HasMask(PermRead|PermUpdate))
	assert.False(t, crud.HasMask(PermRead|PermExport))
	assert.True(t, crud.HasAny(PermExport|PermCreate))

	editor := PermRead | PermCreate | PermUpdate | PermEnable | PermClose
	assert.Equal(t, Mask(55), editor)
	assert.False(t, editor.Has(ActionDelete))

	merged := editor | PermDelete
	assert.True(t, merged.Has(ActionDelete))
	assert.Equal(t, Mask(63), merged)
}

func TestMask_Mutate(t *testing.T) {
	m := PermReadOnly.Add(ActionCreate)
	assert.True(t, m.Has(ActionCreate))
	m = m.Remove(ActionRead)
	assert.False(t, m.Has(ActionRead))
	m = m.Toggle(ActionUpdate)
	assert.True(t, m.Has(ActionUpdate))
}

func TestParseActionAndCodes(t *testing.T) {
	a, ok := ParseAction("DELETE")
	assert.True(t, ok)
	assert.Equal(t, ActionDelete, a)
	_, ok = ParseAction("NOPE")
	assert.False(t, ok)
	assert.Contains(t, PermCRUD.Codes(), "CREATE")
	assert.Contains(t, PermCRUD.Names(), "Add")
}

func TestPermAll(t *testing.T) {
	assert.Equal(t, Mask(1023), PermAll)
	assert.True(t, PermAll.Has(ActionPublish))
}
