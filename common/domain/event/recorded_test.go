package event_test

import (
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/Crows-Storm/Axis/common/domain/user/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromEnvelopes_MapsUserCreated(t *testing.T) {
	msg := events.NewUserCreatedMessage(42, "login", "a@b.c", "alice")
	now := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	got := event.FromEnvelopes([]event.Envelope[event.Message]{
		{Message: msg, Timestamp: now, Metadata: event.Metadata{"k": "v"}},
	})
	require.Len(t, got, 1)
	assert.Equal(t, "UserCreated", got[0].EventName())
	assert.Equal(t, "42", got[0].AggregateID())
	assert.Equal(t, 1, got[0].EventVersion())
	assert.Equal(t, now, got[0].OccurredAt())
	assert.Equal(t, "login", got[0].Payload()["LoginID"])
	assert.Equal(t, "v", got[0].Metadata()["k"])
}

func TestFromEnvelopes_SkipsNilMessage(t *testing.T) {
	got := event.FromEnvelopes([]event.Envelope[event.Message]{{}})
	assert.Empty(t, got)
}
