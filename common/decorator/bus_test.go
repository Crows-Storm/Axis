package decorator

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/stretchr/testify/require"
)

type sampleMessage struct {
	ID          string
	AggregateID uint64
	Category    event.EventCategory
}

func (sampleMessage) Name() string   { return "UserCreated" }
func (sampleMessage) Domain() string { return "user" }

type capturePublisher struct {
	records []Record
	err     error
	closed  bool
}

func (c *capturePublisher) Publish(_ context.Context, records ...Record) error {
	c.records = append(c.records, records...)
	return c.err
}

func (c *capturePublisher) Close() error {
	c.closed = true
	return nil
}

func TestBusMirrorsAuditTopicAndKeepsFlakeID(t *testing.T) {
	pub := &capturePublisher{}
	bus := NewBus(pub, "user", event.AuditTopic)
	ctx := contextx.NewBuilder().WithCorrelationID("corr-1").Build(context.Background())
	id := uint64(639752501632958465)
	err := bus.Publish(ctx, event.Envelope[event.Message]{
		Topic: "axis.user.UserCreated",
		Message: sampleMessage{
			ID:          "evt-1",
			AggregateID: id,
			Category:    event.Created,
		},
		Timestamp: time.Date(2026, 10, 3, 1, 2, 3, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Len(t, pub.records, 2)
	require.Equal(t, "axis.user.UserCreated", pub.records[0].Topic)
	require.Equal(t, event.AuditTopic, pub.records[1].Topic)
	require.Equal(t, "axis.user.UserCreated", pub.records[1].Headers[event.MetaSourceTopic])
	require.Equal(t, "user", pub.records[0].Headers[event.MetaSourceService])
	require.Equal(t, "corr-1", pub.records[0].Headers[contextx.MetadataCorrelationID])
	_, hasCause := pub.records[0].Headers[contextx.MetadataCausationId]
	require.False(t, hasCause)
	require.Equal(t, []byte("639752501632958465"), pub.records[0].Key)

	var doc wireDocument
	require.NoError(t, json.Unmarshal(pub.records[0].Value, &doc))
	require.Equal(t, "evt-1", doc.EventID)
	require.Equal(t, "639752501632958465", doc.AggregateID)
	require.Contains(t, string(doc.Payload), "639752501632958465")
}

func TestBusSkipsAuditCopyWhenTopicMatches(t *testing.T) {
	pub := &capturePublisher{}
	bus := NewBus(pub, "audit", event.AuditTopic)
	err := bus.Publish(context.Background(), event.Envelope[event.Message]{
		Topic:   event.AuditTopic,
		Message: sampleMessage{ID: "evt-2", AggregateID: 1},
	})
	require.NoError(t, err)
	require.Len(t, pub.records, 1)
	require.Equal(t, event.AuditTopic, pub.records[0].Topic)
}

func TestBusDoesNotPublishWhenContextIsCanceled(t *testing.T) {
	pub := &capturePublisher{}
	bus := NewBus(pub, "user", event.AuditTopic)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := bus.Publish(ctx, event.Envelope[event.Message]{
		Topic:   "axis.user.UserCreated",
		Message: sampleMessage{ID: "evt-3"},
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, pub.records)
}

func TestBusReturnsPublisherError(t *testing.T) {
	pub := &capturePublisher{err: errors.New("broker down")}
	bus := NewBus(pub, "user", "")
	err := bus.Publish(context.Background(), event.Envelope[event.Message]{
		Topic:   "axis.user.UserCreated",
		Message: sampleMessage{ID: "evt-4"},
	})
	require.ErrorContains(t, err, "broker down")
}

func TestBusRejectsEmptyTopic(t *testing.T) {
	pub := &capturePublisher{}
	bus := NewBus(pub, "user", event.AuditTopic)
	err := bus.Publish(context.Background(), event.Envelope[event.Message]{
		Message: sampleMessage{ID: "evt-5"},
	})
	require.ErrorContains(t, err, "no topic")
	require.Empty(t, pub.records)
}

func TestBusCopiesEnvelopeCausationMetadata(t *testing.T) {
	pub := &capturePublisher{}
	bus := NewBus(pub, "user", "")
	err := bus.Publish(context.Background(), event.WithCause(
		event.Envelope[event.Message]{
			Topic:   "axis.user.Child",
			Message: sampleMessage{ID: "evt-child", AggregateID: 1},
		},
		event.Wire{
			EventID: "evt-parent",
			Metadata: event.Metadata{
				contextx.MetadataCorrelationID: "corr-chain",
			},
		},
	))
	require.NoError(t, err)
	require.Len(t, pub.records, 1)
	require.Equal(t, "corr-chain", pub.records[0].Headers[contextx.MetadataCorrelationID])
	require.Equal(t, "evt-parent", pub.records[0].Headers[contextx.MetadataCausationId])
}

func TestDecodeWire(t *testing.T) {
	body := []byte(`{"event_id":"evt-1","event_name":"UserCreated","aggregate_id":"7","occurred_at":"2026-10-03T01:02:03Z","version":1,"payload":{"Category":"CREATED"}}`)
	wire, err := DecodeWire("axis.audit", 2, 9, body, map[string]string{
		contextx.MetadataCorrelationID: "corr",
		event.MetaSourceService:        "user",
	})
	require.NoError(t, err)
	require.Equal(t, "evt-1", wire.EventID)
	require.Equal(t, "user", wire.Metadata[event.MetaSourceService])
	require.Equal(t, int64(9), wire.Offset)
}
