package audit

import (
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReadsProducerTrace(t *testing.T) {
	occurred := time.Date(2026, 9, 30, 1, 2, 3, 0, time.UTC)
	rec, err := New(event.Wire{
		EventID:     "evt-1",
		EventName:   "UserCreated",
		AggregateID: "7",
		OccurredAt:  occurred.Format(time.RFC3339Nano),
		Payload:     []byte(`{"Category":"CREATED","OperatorID":9}`),
		Metadata: event.Metadata{
			contextx.MetadataCorrelationID: "corr",
			contextx.MetadataCausationId:   "cause",
			event.MetaSourceService:        "user",
		},
	}, 0)
	require.NoError(t, err)
	assert.Equal(t, "evt-1", rec.EventID)
	assert.Equal(t, uint64(9), rec.OperatorID)
	assert.Equal(t, "CREATED", rec.Category)
	assert.Equal(t, occurred, rec.Timestamp)
	assert.Equal(t, "user", rec.Metadata.Service)
	assert.Equal(t, "corr", rec.Metadata.CorrelationId)
	assert.Equal(t, "cause", rec.Metadata.CausationId)
	assert.Equal(t, "7", rec.Metadata.AggregateID)
}

func TestNewPrefersExplicitOperator(t *testing.T) {
	rec, err := New(event.Wire{
		EventID:   "evt-2",
		EventName: "UserCreated",
		Payload:   []byte(`{"OperatorID":9}`),
	}, 4)
	require.NoError(t, err)
	assert.Equal(t, uint64(4), rec.OperatorID)
}

func TestNewRejectsMissingIdentity(t *testing.T) {
	_, err := New(event.Wire{EventName: "UserCreated"}, 0)
	assert.ErrorIs(t, err, ErrInvalidRecord)
}
