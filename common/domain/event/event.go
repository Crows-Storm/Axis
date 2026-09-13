package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/util"
)

type EventCategory string

// The EventStore is mongo Document, collection name: domain_event
type EventStore struct {
}

const (
	Created EventCategory = "CREATED"
	Updated EventCategory = "UPDATED"
	Deleted EventCategory = "DELETED"

	Activated EventCategory = "ACTIVATED"
	Disabled  EventCategory = "DISABLED"

	Custom EventCategory = "CUSTOM"
	Mark   EventCategory = "MARK"
)

type DomainEvent interface {
	EventID() string
	EventName() string
	EventCategory() EventCategory
	OccurredAtTime() int64

	AggregateId() int64
	AggregateName() string
	EventPayload() json.RawMessage
	EventMetadata() Metadata
}

type BaseEvent struct {
	ID         string        `json:"id,omitempty"` // event unique id, not aggregate id
	Name       string        `json:"name,omitempty"`
	Category   EventCategory `json:"category,omitempty"`
	OccurredAt int64         `json:"occurredAt,omitempty"`

	AggId   int64  `json:"agg_id,omitempty"`
	AggName string `json:"agg_name,omitempty"`

	Payload  json.RawMessage `json:"payload" json:"payload,omitempty"` // Serialized JSON
	Metadata Metadata        `json:"metadata" json:"metadata"`         // store: traceId, correlationId, causationId, operator
}

// Getter for BaseEvent

func (b *BaseEvent) EventID() string               { return b.ID }
func (b *BaseEvent) EventName() string             { return b.Name }
func (b *BaseEvent) EventCategory() EventCategory  { return b.Category }
func (b *BaseEvent) OccurredAtTime() int64         { return b.OccurredAt }
func (b *BaseEvent) AggregateId() int64            { return b.AggId }
func (b *BaseEvent) AggregateName() string         { return b.AggName }
func (b *BaseEvent) EventPayload() json.RawMessage { return b.Payload }
func (b *BaseEvent) EventMetadata() Metadata       { return b.Metadata }

type Metadata struct {
	TraceId string `json:"trace_id"`
	SpanId  string `json:"span_id"`
	CorrID  string `json:"correlation_id"`         // Operation-level correlation ID
	CauseID string `json:"causation_id,omitempty"` // Causation chain ID (the source event ID that triggered this event)
}

// The EventMetadataWithContext TODO: need Implement it!
func EventMetadataWithContext(ctx context.Context) Metadata {
	return Metadata{
		TraceId: "",
		SpanId:  "",
		CorrID:  "",
		CauseID: "",
	}
}

func WithCorrelationId(ctx context.Context, correlationId string) context.Context {
	return context.WithValue(ctx, "correlation_id", correlationId)
}

func CorrelationIdFromContext(ctx context.Context) string {
	return ctx.Value("correlation_id").(string)
}

type MetadataBuilder struct {
	metadata *Metadata
}

func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{metadata: &Metadata{}}
}

func (b *MetadataBuilder) WithTraceId(id string) *MetadataBuilder {
	b.metadata.TraceId = id
	return b
}

// TODO: need get from context
func (b *MetadataBuilder) WithCorrelationId(id string) *MetadataBuilder {
	if id == "" {
		id = generateCorrelationID()
	}
	b.metadata.CorrID = id
	return b
}

// TODO: need get from context
func (b *MetadataBuilder) WithCausationId(id string) *MetadataBuilder {
	if id == "" {
		id = generateCausationID()
	}
	return b
}

func (b *MetadataBuilder) Build() *Metadata {
	return b.metadata
}

// Usage:
// m1 := NewMetadata("trace-123")                           // Only TraceId
// m2 := NewMetadata("trace-123", "corr-456")              // TraceId + CorrID
// m3 := NewMetadata("trace-123", "corr-456", "cause-789") // All three

func NewBaseEvent(name string, category EventCategory, aggregateId int64, aggregateName string) BaseEvent {
	return BaseEvent{
		ID:         util.UUIDV1().String(),
		Name:       name,
		Category:   category,
		OccurredAt: time.Now().UnixMilli(), // e.g. 1788601067064

		AggId:   aggregateId,
		AggName: aggregateName,

		//Payload: json.RawMessage(`{}`),
		//Metadata: Metadata{},
	}
}

type EventBuilder struct {
	id         string // event unique id, not aggregate id
	name       string
	category   EventCategory
	occurredAt int64

	aggId   int64
	aggName string

	payload  json.RawMessage // Serialized JSON
	metadata Metadata

	err error
}

func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		payload: json.RawMessage{},
		metadata: Metadata{
			TraceId: "",
			CorrID:  "", // Operation-level correlation ID
			CauseID: "",
		},
	}
}

func (b *EventBuilder) WithName(name string) *EventBuilder {
	if b.err != nil {
		return b
	}
	if name == "" {
		b.err = fmt.Errorf("event name cannot be empty")
		return b
	}
	b.name = name
	return b
}

func (b *EventBuilder) WithEventCategory(category EventCategory) *EventBuilder {
	if b.err != nil {
		return b
	}
	if category == "" {
		b.err = fmt.Errorf("event category cannot be empty")
		return b
	}
	b.category = category
	return b
}

func (b *EventBuilder) WithAggregate(aggregateID int64, aggregateName string) *EventBuilder {
	if b.err != nil {
		return b
	}
	if aggregateID <= 0 {
		b.err = fmt.Errorf("aggregate ID must be positive")
		return b
	}
	if aggregateName == "" {
		b.err = fmt.Errorf("aggregate name cannot be empty")
		return b
	}
	b.aggId = aggregateID
	b.aggName = aggregateName
	return b
}

func (b *EventBuilder) WithPayload(data interface{}) *EventBuilder {
	if b.err != nil {
		return b
	}
	if data == nil {
		b.err = fmt.Errorf("event data cannot be nil")
		return b
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		b.err = fmt.Errorf("failed to marshal event data: %w", err)
		return b
	}
	b.payload = payloadBytes
	return b
}

func (b *EventBuilder) WithMetadata(metadata Metadata) *EventBuilder {
	if b.err != nil {
		return b
	}
	if metadata.TraceId != "" {
		b.metadata.TraceId = metadata.TraceId
	}
	if metadata.CorrID != "" {
		b.metadata.CorrID = metadata.CorrID
	}
	if metadata.CauseID != "" {
		b.metadata.CauseID = metadata.CauseID
	}
	return b
}

func (b *EventBuilder) WithTraceID(traceID string) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.metadata.TraceId = traceID
	return b
}

func (b *EventBuilder) WithCorrelationID(corrID string) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.metadata.CorrID = corrID
	return b
}

func (b *EventBuilder) WithCausationID(causeID string) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.metadata.CauseID = causeID
	return b
}

// The Build is build a new DomainEvent, so new a event id in itself
func (b *EventBuilder) Build() (DomainEvent, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.id == "" {
		b.id = generateCorrelationID()
	}
	if b.name == "" {
		return nil, fmt.Errorf("event name is required")
	}
	if b.category == "" {
		return nil, fmt.Errorf("event category is required")
	}
	//if b.occurredAt <= 0 {
	//	return nil, fmt.Errorf("occurredAt is required")
	//}
	if b.aggId <= 0 {
		return nil, fmt.Errorf("aggregate ID is required")
	}
	if b.aggName == "" {
		return nil, fmt.Errorf("aggregate name is required")
	}
	if len(b.payload) == 0 {
		return nil, fmt.Errorf("event payload is required")
	}

	return &BaseEvent{
		ID:         generateEventID(),
		Name:       b.name,
		Category:   b.category,
		OccurredAt: time.Now().UnixMilli(),

		AggId:   b.aggId,
		AggName: b.aggName,

		Payload:  b.payload,
		Metadata: b.metadata,
	}, nil
}

func (*EventBuilder) generateEventID() string {
	return util.UUIDV1().String()
}

func generateEventID() string {
	// Generate random UUID
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), util.UUIDV4())
}

func generateCorrelationID() string {
	return util.UUIDV4().String()
}

func generateCausationID() string {
	return util.UUIDV4().String()
}

type EventHandler interface {
	EventName() string
	Handle(ctx context.Context, envelope Envelope) error
}
