package kafka

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// Serializer 负责领域事件的序列化/反序列化
type Serializer struct {
	mu       sync.RWMutex
	registry map[string]eventFactory // eventName -> factory
}

type eventFactory func() event.DomainEvent

func NewSerializer() *Serializer {
	return &Serializer{
		registry: make(map[string]eventFactory),
	}
}

// Register 注册事件类型（用于反序列化）
func (s *Serializer) Register(eventName string, factory eventFactory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry[eventName] = factory
}

// Serialize 将领域事件序列化为 JSON
func (s *Serializer) Serialize(evt event.DomainEvent) ([]byte, error) {
	envelope := MessageEnvelope{
		EventID:     uuid.New().String(),
		EventName:   evt.EventName(),
		AggregateID: evt.AggregateID(),
		OccurredAt:  evt.OccurredAt().Format(time.RFC3339Nano),
		Version:     evt.EventVersion(),
		Payload:     evt.Payload(),
		Metadata:    evt.Metadata(),
	}
	return json.Marshal(envelope)
}

// Deserialize 从 JSON 反序列化为 Envelope
func (s *Serializer) Deserialize(data []byte) (*event.Envelope, error) {
	var env MessageEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(env.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	return &event.Envelope{
		EventID:     env.EventID,
		EventName:   env.EventName,
		AggregateID: env.AggregateID,
		OccurredAt:  env.OccurredAt,
		Version:     env.Version,
		Payload:     payloadBytes,
		Metadata:    env.Metadata,
	}, nil
}

// DeserializeEnvelope 从 Kafka 消息解析 Envelope
func DeserializeEnvelope(msg *sarama.Message) (*event.Envelope, error) {
	envelope, err := NewSerializer().Deserialize(msg.Value)
	if err != nil {
		return nil, err
	}
	// 从 headers 补充信息
	for _, h := range msg.Headers {
		key := string(h.Key)
		val := string(h.Value)
		if envelope.Metadata == nil {
			envelope.Metadata = make(map[string]string)
		}
		envelope.Metadata[key] = val
	}
	return envelope, nil
}

// MessageEnvelope 传输层信封结构
type MessageEnvelope struct {
	EventID     string                 `json:"event_id"`
	EventName   string                 `json:"event_name"`
	AggregateID string                 `json:"aggregate_id"`
	OccurredAt  string                 `json:"occurred_at"`
	Version     int                    `json:"version"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}
