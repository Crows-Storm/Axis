package event_test

import (
	"testing"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/stretchr/testify/assert"
)

type causeMsg struct{}

func (causeMsg) Name() string   { return "Child" }
func (causeMsg) Domain() string { return "user" }

func TestWithCause_CopiesCorrelationAndParentEventID(t *testing.T) {
	env := event.Envelope[event.Message]{
		Topic:   "axis.user.Child",
		Message: causeMsg{},
	}
	got := event.WithCause(env, event.Wire{
		EventID: "evt-parent",
		Metadata: event.Metadata{
			contextx.MetadataCorrelationID: "corr-1",
		},
	})
	assert.Equal(t, "corr-1", got.Metadata[contextx.MetadataCorrelationID])
	assert.Equal(t, "evt-parent", got.Metadata[contextx.MetadataCausationId])
}

func TestWithCause_FallsBackToHeaders(t *testing.T) {
	got := event.WithCause(event.Envelope[event.Message]{Message: causeMsg{}}, event.Wire{
		EventID: "evt-2",
		Headers: map[string]string{contextx.MetadataCorrelationID: "corr-h"},
	})
	assert.Equal(t, "corr-h", got.Metadata[contextx.MetadataCorrelationID])
	assert.Equal(t, "evt-2", got.Metadata[contextx.MetadataCausationId])
}
