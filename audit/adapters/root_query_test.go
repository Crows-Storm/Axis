package adapters

import (
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/audit/domain/audit"
	"github.com/Crows-Storm/Axis/common/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRootQueryRequiresCorrelationAndEmptyCausation(t *testing.T) {
	query, err := rootQuery(audit.RootFilter{}, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(query), 2)
	assert.Equal(t, fieldCorrelationID, query[0].Key)
	assert.Equal(t, fieldCausationID, query[1].Key)
}

func TestRootQueryAppliesWindowAndSeek(t *testing.T) {
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	seek := &pagination.Seek{Value: from.Format(time.RFC3339Nano), ID: "evt-9", Greater: false}
	query, err := rootQuery(audit.RootFilter{Service: "user", EventName: "UserCreated", From: from, To: to}, seek)
	require.NoError(t, err)

	keys := make([]string, 0, len(query))
	for _, e := range query {
		keys = append(keys, e.Key)
	}
	assert.Contains(t, keys, fieldService)
	assert.Contains(t, keys, fieldEventName)
	assert.Contains(t, keys, fieldTimestamp)
	assert.Contains(t, keys, "$or")
}

func TestRootQueryRejectsBadCursorTime(t *testing.T) {
	_, err := rootQuery(audit.RootFilter{}, &pagination.Seek{Value: "not-a-time", ID: "x"})
	require.Error(t, err)
}

func TestMongoSeekMergesAsDocumentKeys(t *testing.T) {
	filter := pagination.MongoSeekFilter(fieldTimestamp, fieldEventID, &pagination.Seek{ID: "e", Greater: true}, time.Now().UTC())
	require.Equal(t, "$or", filter[0].Key)
	_, ok := filter[0].Value.(bson.A)
	assert.True(t, ok)
}
