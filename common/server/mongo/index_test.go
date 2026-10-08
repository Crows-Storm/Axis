package mongo

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestIndexModelsDirections(t *testing.T) {
	models, err := IndexModels([]*IndexInfo{
		{Name: "event_id", Fields: []string{"event_id"}, Type: ASC},
		{Name: "metadata_timestamp", Fields: []string{" metadata.service ", "timestamp"}, Type: DESC},
	})
	require.NoError(t, err)
	require.Equal(t, bson.D{{Key: "event_id", Value: int32(1)}}, models[0].Keys)
	require.Equal(t, bson.D{
		{Key: "metadata.service", Value: int32(-1)},
		{Key: "timestamp", Value: int32(-1)},
	}, models[1].Keys)
}

func TestIndexModelsRejectsInvalidSpecs(t *testing.T) {
	_, err := IndexModels([]*IndexInfo{nil})
	require.Error(t, err)
	_, err = IndexModels([]*IndexInfo{{Fields: []string{"event_id"}}})
	require.Error(t, err)
	_, err = IndexModels([]*IndexInfo{{Name: "event_id", Fields: []string{"event_id"}, Type: "text"}})
	require.Error(t, err)
	_, err = IndexModels([]*IndexInfo{
		{Name: "event_id", Fields: []string{"event_id"}},
		{Name: "event_id", Fields: []string{"aggregate_id"}},
	})
	require.Error(t, err)
}
