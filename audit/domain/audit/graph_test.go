package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForestSiblingRoots(t *testing.T) {
	ts := time.Date(2026, 10, 4, 1, 0, 0, 0, time.UTC)
	created := &AuditEvent{EventID: "c", Timestamp: ts, Metadata: Metadata{CorrelationId: "corr"}}
	apply := &AuditEvent{EventID: "p", Timestamp: ts.Add(time.Second), Metadata: Metadata{CorrelationId: "corr"}}
	child := &AuditEvent{
		EventID:   "k",
		Timestamp: ts.Add(2 * time.Second),
		Metadata:  Metadata{CorrelationId: "corr", CausationId: "c"},
	}
	graph := Forest(created, []*AuditEvent{created, apply, child})
	require.Len(t, graph.Roots, 2)
	assert.Equal(t, "c", graph.Roots[0].Event.EventID)
	require.Len(t, graph.Roots[0].Children, 1)
	assert.Equal(t, "k", graph.Roots[0].Children[0].Event.EventID)
	assert.Equal(t, "p", graph.Roots[1].Event.EventID)
}

func TestForestWithoutCorrelation(t *testing.T) {
	ev := &AuditEvent{EventID: "solo"}
	graph := Forest(ev, nil)
	require.Len(t, graph.Roots, 1)
	assert.Equal(t, "solo", graph.Roots[0].Event.EventID)
}

func TestIsRoot(t *testing.T) {
	assert.True(t, (&AuditEvent{Metadata: Metadata{CorrelationId: "c"}}).IsRoot())
	assert.False(t, (&AuditEvent{Metadata: Metadata{CorrelationId: "c", CausationId: "p"}}).IsRoot())
	assert.False(t, (&AuditEvent{}).IsRoot())
}
