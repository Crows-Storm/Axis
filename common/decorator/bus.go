package decorator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/event"
)

// Record is one transport message. The Kafka client sends these bytes unchanged.
type Record struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers map[string]string
}

// Publisher is the transport port behind EventBus.
type Publisher interface {
	Publish(ctx context.Context, records ...Record) error
	Close() error
}

// Bus encodes domain envelopes and optionally copies each record onto an audit topic.
type Bus struct {
	publisher   Publisher
	serviceName string
	auditTopic  string
}

// NewBus builds the glue between aggregates and a transport publisher.
// The business topic is Envelope.Topic. auditTopic receives a copy when it is set
// and differs from that topic.
func NewBus(publisher Publisher, serviceName, auditTopic string) *Bus {
	return &Bus{
		publisher:   publisher,
		serviceName: serviceName,
		auditTopic:  auditTopic,
	}
}

func (b *Bus) Publish(ctx context.Context, events ...event.Envelope[event.Message]) error {
	if b == nil || b.publisher == nil || len(events) == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	records, err := b.encode(ctx, events)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	if err := b.publisher.Publish(ctx, records...); err != nil {
		return fmt.Errorf("publish events: %w", err)
	}
	return nil
}

func (b *Bus) Close() error {
	if b == nil || b.publisher == nil {
		return nil
	}
	return b.publisher.Close()
}

func (b *Bus) encode(ctx context.Context, events []event.Envelope[event.Message]) ([]Record, error) {
	// Trace headers come from context. Domain causation is only on Envelope.Metadata.
	headers := contextx.InjectIntoMQHeaders(ctx)
	if b.serviceName != "" {
		headers[event.MetaSourceService] = b.serviceName
	}
	records := make([]Record, 0, len(events)*2)
	version := 0
	for _, env := range events {
		if env.Message == nil {
			continue
		}
		version++
		topic := env.Topic
		if topic == "" {
			return nil, fmt.Errorf("event %s has no topic", env.Message.Name())
		}
		body, key, err := encodeEnvelope(env, version)
		if err != nil {
			return nil, err
		}
		recHeaders := copyHeaders(headers)
		for k, v := range env.Metadata {
			if v != "" {
				recHeaders[k] = v
			}
		}
		recHeaders["event_name"] = env.Message.Name()
		records = append(records, Record{Topic: topic, Key: key, Value: body, Headers: recHeaders})
		if b.auditTopic != "" && b.auditTopic != topic {
			auditHeaders := copyHeaders(recHeaders)
			auditHeaders[event.MetaSourceTopic] = topic
			records = append(records, Record{Topic: b.auditTopic, Key: key, Value: body, Headers: auditHeaders})
		}
	}
	return records, nil
}

type wireDocument struct {
	EventID     string          `json:"event_id"`
	EventName   string          `json:"event_name"`
	AggregateID string          `json:"aggregate_id"`
	OccurredAt  string          `json:"occurred_at"`
	Version     int             `json:"version"`
	Payload     json.RawMessage `json:"payload"`
}

func encodeEnvelope(env event.Envelope[event.Message], version int) (body, key []byte, err error) {
	payload, err := json.Marshal(env.Message)
	if err != nil {
		return nil, nil, fmt.Errorf("encode %s: %w", env.Message.Name(), err)
	}
	eventID, aggregateID := identityOf(payload)
	if eventID == "" {
		eventID = event.GenerateEventID()
	}
	occurred := env.Timestamp
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}
	raw, err := json.Marshal(wireDocument{
		EventID:     eventID,
		EventName:   env.Message.Name(),
		AggregateID: aggregateID,
		OccurredAt:  occurred.UTC().Format(time.RFC3339Nano),
		Version:     version,
		Payload:     payload,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("encode %s: %w", env.Message.Name(), err)
	}
	if aggregateID != "" {
		key = []byte(aggregateID)
	}
	return raw, key, nil
}

func identityOf(payload []byte) (eventID, aggregateID string) {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var probe map[string]any
	if err := dec.Decode(&probe); err != nil {
		return "", ""
	}
	if id, ok := probe["ID"].(string); ok {
		eventID = id
	}
	switch n := probe["AggregateID"].(type) {
	case string:
		aggregateID = n
	case json.Number:
		aggregateID = n.String()
	}
	return eventID, aggregateID
}

// DecodeWire reads a consumed record back into the audit contract.
func DecodeWire(topic string, partition int, offset int64, value []byte, headers map[string]string) (event.Wire, error) {
	var doc wireDocument
	if err := json.Unmarshal(value, &doc); err != nil {
		return event.Wire{}, fmt.Errorf("decode event: %w", err)
	}
	if doc.EventID == "" || doc.EventName == "" {
		return event.Wire{}, fmt.Errorf("decode event: event_id and event_name are required")
	}
	md := event.Metadata{}
	for k, v := range headers {
		if v != "" {
			md[k] = v
		}
	}
	return event.Wire{
		EventID:     doc.EventID,
		EventName:   doc.EventName,
		AggregateID: doc.AggregateID,
		OccurredAt:  doc.OccurredAt,
		Version:     doc.Version,
		Payload:     doc.Payload,
		Metadata:    md,
		Headers:     headers,
		Topic:       topic,
		Partition:   partition,
		Offset:      offset,
	}, nil
}

func copyHeaders(src map[string]string) map[string]string {
	out := make(map[string]string, len(src)+2)
	for k, v := range src {
		out[k] = v
	}
	return out
}
