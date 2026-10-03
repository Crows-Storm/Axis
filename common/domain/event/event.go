package event

import (
	"context"
	"maps"
	"time"

	"github.com/Crows-Storm/Axis/common/util"
)

type Message interface {
	Name() string
}

type Envelope[T Message] struct {
	Topic string
	Key   []byte
	Value []byte

	Partition int
	Offset    int64
	Timestamp time.Time

	Message T
	// The Metadata is
	//MetadataRequestID     = "x-request-id"
	//MetadataCorrelationID = "x-correlation-id"
	//MetadataCausationId   = "x-causation-id"
	//MetadataTraceID       = "x-trace-id"
	//MetadataSpanID        = "x-span-id"
	//MetadataParentSpanID  = "x-parent-span-id"
	Metadata Metadata

	// Reserved: Original underlying message reference
	raw any
}

// Metadata contains some data related to a Message that are not functional
// for the Message itself, but instead functioning as supporting information
// to provide additional context.
type Metadata map[string]string

// With returns a new Metadata reference holding the value addressed using
// the specified key.
func (m Metadata) With(key, value string) Metadata {
	if m == nil {
		m = make(Metadata)
	}
	m[key] = value
	return m
}

// Merge merges the other Metadata provided in input with the current map.
// Returns a pointer to the extended metadata map.
func (m Metadata) Merge(other Metadata) Metadata {
	if m == nil {
		return other
	}
	maps.Copy(m, other)
	return m
}

type GenericEnvelope Envelope[Message]

const (
	// MetaSourceService names the producing process. It is not a principal.
	MetaSourceService = "source_service"
	// MetaSourceTopic is the business topic copied onto an audit message.
	MetaSourceTopic = "source_topic"
)

type EventCategory string

const (
	Created EventCategory = "CREATED"
	Updated EventCategory = "UPDATED"
	Deleted EventCategory = "DELETED"

	ApplyPassword EventCategory = "APPLY_PASSWORD"

	Activated EventCategory = "ACTIVATED"
	Disabled  EventCategory = "DISABLED"

	Custom EventCategory = "CUSTOM"
	Mark   EventCategory = "MARK"
)

// ToGenericEnvelope maps the Envelope instance into a GenericEnvelope one.
func (e Envelope[T]) ToGenericEnvelope() GenericEnvelope {
	return GenericEnvelope{
		Message:  e.Message,
		Metadata: e.Metadata,
	}
}

// The GenerateEventID to generate a EventID
func GenerateEventID() string {
	return util.UUIDV4().String()
}

// DomainEvent is the publish-side contract. Identity lives on the event;
// request trace lives on context and is projected to MQ headers separately.
type DomainEvent interface {
	EventName() string
	EventVersion() int
	AggregateID() string
	OccurredAt() time.Time
	Payload() map[string]interface{}
	Metadata() Metadata
}

// Wire is the consume-side envelope. Headers/Metadata carry contextx MQ keys.
type Wire struct {
	EventID     string
	EventName   string
	AggregateID string
	OccurredAt  string
	Version     int
	Payload     []byte
	Metadata    Metadata
	Headers     map[string]string
	Topic       string
	Partition   int
	Offset      int64
}

type Dispatcher interface {
	Dispatch(ctx context.Context, env Wire) error
}
