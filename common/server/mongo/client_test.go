package mongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type stubDoc struct {
	name string
	ts   bool
}

func (s stubDoc) CollectionName() string { return s.name }
func (s stubDoc) IsTimeSeries() bool     { return s.ts }

func TestCollectionsToCreateSkipsExisting(t *testing.T) {
	docs := []Document{
		stubDoc{name: "audit_event", ts: true},
		stubDoc{name: " audit_event ", ts: true},
		stubDoc{name: "audit_meta", ts: false},
	}
	pending, err := collectionsToCreate(docs, []string{"audit_event"})
	require.NoError(t, err)
	require.Equal(t, []Document{stubDoc{name: "audit_meta", ts: false}}, pending)
}

func TestCollectionsToCreateRejectsConflicts(t *testing.T) {
	_, err := collectionsToCreate([]Document{
		stubDoc{name: "audit_event", ts: true},
		stubDoc{name: "audit_event", ts: false},
	}, nil)
	require.Error(t, err)

	_, err = collectionsToCreate([]Document{nil}, nil)
	require.Error(t, err)

	_, err = collectionsToCreate([]Document{stubDoc{}}, nil)
	require.Error(t, err)
}

func TestInitCollectionRejectsNilClient(t *testing.T) {
	var client *Client
	require.Error(t, client.InitCollection(t.Context(), stubDoc{name: "audit_event", ts: true}))
}
