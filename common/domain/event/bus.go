package event

import (
	"context"
	"encoding/json"

	"github.com/Crows-Storm/Axis/common/config/logger"
)

type Bus interface {
	Publish(ctx context.Context, event DomainEvent) error
	PublishAll(ctx context.Context, events []DomainEvent) error
	//Subscribe(eventName string, handler EventHandler) error
	//Unsubscribe(eventName string, handler EventHandler) error

	CloseAll() error
}

type Envelope struct {
	EventID       string `json:"event_id"`
	EventName     string `json:"event_name"`
	AggregateID   string `json:"aggregate_id"`
	AggregateName string `json:"aggregate_name"`
	OccurredAt    string `json:"occurred_at"`

	Payload  json.RawMessage   `json:"payload"`
	Metadata map[string]string `json:"metadata"`

	Topic      string `json:"-"`
	Partition  int    `json:"-"`
	Offset     int64  `json:"-"`
	RetryCount int    `json:"-"`
}

// The ToEnvelope is factory method for Envelope
// TODO Here need optimization
func ToEnvelope(event DomainEvent) Envelope {
	if event == nil {
		logger.Warn("nil event")
		return Envelope{}
	}
	return buildEnvelopes(event)
}

// The ToEnvelopes is factory method for Envelope
func ToEnvelopes(events ...DomainEvent) []Envelope {
	envelopes := make([]Envelope, len(events))
	for _, event := range events {
		if event == nil {
			logger.Warn("nil event")
			envelopes = append(envelopes, Envelope{})
		}
		envelopes = append(envelopes, buildEnvelopes(event))
	}
	return envelopes
}

func buildEnvelopes(event DomainEvent) Envelope {
	return Envelope{
		EventID:       "",
		EventName:     "",
		AggregateID:   "",
		AggregateName: "",
		OccurredAt:    "",

		Payload:  event.EventPayload(),
		Metadata: nil,

		Topic:      "",
		Partition:  0,
		Offset:     0,
		RetryCount: 0,
	}
}
