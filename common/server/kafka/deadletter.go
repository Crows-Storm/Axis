package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// DeadLetterProducer 死信队列生产者
type DeadLetterProducer struct {
	producer *Producer
	topic    string
	logger   zerolog.Logger
}

func NewDeadLetterProducer(producer *Producer, topic string) *DeadLetterProducer {
	return &DeadLetterProducer{
		producer: producer,
		topic:    topic,
	}
}

// DeadLetterMessage 死信消息结构
type DeadLetterMessage struct {
	OriginalTopic     string            `json:"original_topic"`
	OriginalPartition int32             `json:"original_partition"`
	OriginalOffset    int64             `json:"original_offset"`
	EventID           string            `json:"event_id"`
	EventName         string            `json:"event_name"`
	AggregateID       string            `json:"aggregate_id"`
	Payload           json.RawMessage   `json:"payload"`
	Metadata          map[string]string `json:"metadata"`
	Error             string            `json:"error"`
	FailedAt          time.Time         `json:"failed_at"`
}

func (d *DeadLetterProducer) Send(ctx context.Context, msg *sarama.Message, envelope *event.Envelope, processErr error) error {
	dlm := DeadLetterMessage{
		OriginalTopic:     msg.Topic,
		OriginalPartition: msg.Partition,
		OriginalOffset:    msg.Offset,
		FailedAt:          time.Now(),
	}

	if envelope != nil {
		dlm.EventID = envelope.EventID
		dlm.EventName = envelope.EventName
		dlm.AggregateID = envelope.AggregateID
		dlm.Payload = envelope.Payload
		dlm.Metadata = envelope.Metadata
	}
	if processErr != nil {
		dlm.Error = processErr.Error()
	}

	payload, err := json.Marshal(dlm)
	if err != nil {
		return err
	}

	dlMsg := &sarama.ProducerMessage{
		Topic: d.topic,
		Key:   msg.Key,
		Value: sarama.ByteEncoder(payload),
	}

	_, _, err = d.producer.SendMessage(ctx, dlMsg)
	if err != nil {
		d.logger.Error().Err(err).Msg("Failed to produce dead letter message")
	} else {
		d.logger.Warn().
			Str("event", dlm.EventName).
			Str("error", dlm.Error).
			Msg("Message sent to dead letter queue")
	}
	return err
}
