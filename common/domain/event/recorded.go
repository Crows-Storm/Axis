package event

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type recorded struct {
	name        string
	version     int
	aggregateID string
	occurredAt  time.Time
	payload     map[string]interface{}
	metadata    Metadata
}

func (r recorded) EventName() string               { return r.name }
func (r recorded) EventVersion() int               { return r.version }
func (r recorded) AggregateID() string             { return r.aggregateID }
func (r recorded) OccurredAt() time.Time           { return r.occurredAt }
func (r recorded) Payload() map[string]interface{} { return r.payload }
func (r recorded) Metadata() Metadata              { return r.metadata }

func FromEnvelopes(envelopes []Envelope[Message]) []DomainEvent {
	out := make([]DomainEvent, 0, len(envelopes))
	for i, env := range envelopes {
		if env.Message == nil {
			continue
		}
		out = append(out, fromEnvelope(env, i+1))
	}
	return out
}

func fromEnvelope(env Envelope[Message], version int) DomainEvent {
	occurred := env.Timestamp
	if occurred.IsZero() {
		occurred = time.Now()
	}
	payload := messagePayload(env.Message)
	return recorded{
		name:        env.Message.Name(),
		version:     version,
		aggregateID: aggregateIDFromPayload(payload),
		occurredAt:  occurred,
		payload:     payload,
		metadata:    env.Metadata,
	}
}

func messagePayload(msg Message) map[string]interface{} {
	raw, err := json.Marshal(msg)
	if err != nil {
		return map[string]interface{}{"name": msg.Name()}
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]interface{}{"name": msg.Name()}
	}
	return payload
}

func aggregateIDFromPayload(payload map[string]interface{}) string {
	v, ok := payload["AggregateID"]
	if !ok {
		return ""
	}
	switch n := v.(type) {
	case string:
		return n
	case float64:
		return strconv.FormatInt(int64(n), 10)
	case json.Number:
		return n.String()
	default:
		return fmt.Sprint(n)
	}
}
