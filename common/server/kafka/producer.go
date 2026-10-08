package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// sender is the subset of sarama.SyncProducer used by Send.
// Send waits for this call to finish. It does not abandon the wait when the
// caller context is cancelled, because a late ack would otherwise be reported
// as a failure after the broker had already stored the record.
type sender interface {
	SendMessages(msgs []*sarama.ProducerMessage) error
	Close() error
}

// Producer publishes records and returns the broker result.
type Producer struct {
	sync   sender
	once   sync.Once
	closed error
}

func newProducer(brokers []string, clientID string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_6_0_0
	cfg.ClientID = clientID
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Idempotent = true
	cfg.Net.MaxOpenRequests = 1
	cfg.Producer.Return.Successes = true
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Timeout = 5 * time.Second
	cfg.Net.DialTimeout = 5 * time.Second
	cfg.Net.ReadTimeout = 5 * time.Second
	cfg.Net.WriteTimeout = 5 * time.Second

	syncProducer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}
	return &Producer{sync: syncProducer}, nil
}

// Send returns the producer result. A context that is already cancelled does
// not start a send. Cancellation after the send has started does not change
// the result: the caller waits until the broker ack or the producer error.
func (p *Producer) Send(ctx context.Context, msgs []*sarama.ProducerMessage) error {
	if p == nil || p.sync == nil {
		return fmt.Errorf("kafka producer is nil")
	}
	if len(msgs) == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := p.sync.SendMessages(msgs); err != nil {
		return fmt.Errorf("kafka publish: %w", err)
	}
	return nil
}

// Close closes the underlying producer once.
func (p *Producer) Close() error {
	if p == nil || p.sync == nil {
		return nil
	}
	p.once.Do(func() {
		p.closed = p.sync.Close()
	})
	return p.closed
}
