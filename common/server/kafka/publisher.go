package kafka

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/IBM/sarama"
)

type transport struct {
	producer *Producer
}

func (t *transport) Publish(ctx context.Context, records ...decorator.Record) error {
	if len(records) == 0 {
		return nil
	}
	msgs := make([]*sarama.ProducerMessage, 0, len(records))
	for _, rec := range records {
		if rec.Topic == "" {
			return fmt.Errorf("kafka record has no topic")
		}
		headers := make([]sarama.RecordHeader, 0, len(rec.Headers))
		for k, v := range rec.Headers {
			if k == "" || v == "" {
				continue
			}
			headers = append(headers, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
		}
		msgs = append(msgs, &sarama.ProducerMessage{
			Topic:   rec.Topic,
			Key:     sarama.ByteEncoder(rec.Key),
			Value:   sarama.ByteEncoder(rec.Value),
			Headers: headers,
		})
	}
	return t.producer.Send(ctx, msgs)
}

func (t *transport) Close() error {
	if t == nil || t.producer == nil {
		return nil
	}
	return t.producer.Close()
}
