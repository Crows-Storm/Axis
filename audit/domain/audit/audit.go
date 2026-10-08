package audit

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Crows-Storm/Axis/common/server/mongo"
)

var (
	ErrNotFound      = errors.New("audit event not found")
	ErrInvalidRecord = errors.New("invalid audit event")
)

type AuditEvent struct {
	ID         string    `json:"id" bson:"_id"`
	EventID    string    `json:"event_id" bson:"event_id"`
	EventName  string    `json:"event_name" bson:"event_name"`
	OperatorID uint64    `json:"operator_id" bson:"operator_id"`
	Category   string    `json:"category" bson:"category"`
	Timestamp  time.Time `json:"timestamp" bson:"timestamp"`
	Metadata   Metadata  `json:"metadata" bson:"metadata"`
}

type Metadata struct {
	Service       string          `json:"service" bson:"service"`
	AggregateID   string          `json:"aggregate_id" bson:"aggregate_id"`
	CorrelationId string          `json:"correlation_id" bson:"correlation_id"`
	CausationId   string          `json:"causation_id" bson:"causation_id"`
	Payload       json.RawMessage `json:"payload,omitempty" bson:"payload,omitempty"`
}

func (a *AuditEvent) CollectionName() string {
	return "audit_event"
}

func (a *AuditEvent) IsTimeSeries() bool {
	return true
}

func (a *AuditEvent) IsRoot() bool {
	if a == nil {
		return false
	}
	return strings.TrimSpace(a.Metadata.CorrelationId) != "" && strings.TrimSpace(a.Metadata.CausationId) == ""
}

func (a *AuditEvent) GetCursorValue(field string) string {
	if a == nil {
		return ""
	}
	_ = field
	return a.Timestamp.UTC().Format(time.RFC3339Nano)
}

func (a *AuditEvent) GetID() string {
	if a == nil {
		return ""
	}
	return a.EventID
}

func (a *AuditEvent) Index() []*mongo.IndexInfo {
	return []*mongo.IndexInfo{
		{
			Name:   "index_event_id",
			Fields: []string{"event_id"},
			Type:   mongo.ASC,
		},
		{
			Name:   "index_causation_time",
			Fields: []string{"metadata.causation_id", mongo.TimestampField},
			Type:   mongo.DESC,
		},
		{
			Name:   "index_correlation_time",
			Fields: []string{"metadata.correlation_id", mongo.TimestampField},
			Type:   mongo.DESC,
		},
	}
}
