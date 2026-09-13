package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OutboxMessage 外发信箱消息
type OutboxMessage struct {
	ID          string     `json:"id" db:"id"`
	AggregateID string     `json:"aggregate_id" db:"aggregate_id"`
	EventName   string     `json:"event_name" db:"event_name"`
	Payload     string     `json:"payload" db:"payload"`
	Metadata    string     `json:"metadata" db:"metadata"`
	Status      string     `json:"status" db:"status"` // pending, published, failed
	RetryCount  int        `json:"retry_count" db:"retry_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
	ErrorMsg    *string    `json:"error_msg,omitempty" db:"error_msg"`
}

// OutboxRepository 外发信箱仓储接口
type OutboxRepository interface {
	Save(ctx context.Context, msg *OutboxMessage) error
	SaveBatch(ctx context.Context, msgs []*OutboxMessage) error
	FindPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkPublished(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
	IncrementRetry(ctx context.Context, id string) error
	CleanupOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// NewOutboxMessage 从领域事件创建 outbox 消息
func NewOutboxMessage(eventName, aggregateID string, payload map[string]interface{}, metadata map[string]string) (*OutboxMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	return &OutboxMessage{
		ID:          uuid.New().String(),
		AggregateID: aggregateID,
		EventName:   eventName,
		Payload:     string(payloadBytes),
		Metadata:    string(metaBytes),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}, nil
}
