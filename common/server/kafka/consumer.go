package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/IBM/sarama"
)

// Delivery is one consumed record.
type Delivery struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
}

// Handler processes one delivery. A non-nil error sends the record to the dead-letter topic.
type Handler func(ctx context.Context, delivery Delivery) error

// Group consumes topics until Close or the parent context is cancelled.
type Group struct {
	cancel   context.CancelFunc
	done     chan struct{}
	group    sarama.ConsumerGroup
	producer *Producer
}

// OpenConsumer joins groupID and reads topics from the earliest offset.
func OpenConsumer(parent context.Context, cfg config.KafkaConfig, serviceName string, topics []string, handle Handler) (*Group, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("kafka brokers are empty")
	}
	if handle == nil {
		return nil, fmt.Errorf("kafka handler is nil")
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("kafka topics are empty")
	}
	clientID := firstNonEmpty(cfg.ClientID, serviceName, "axis")
	groupID := firstNonEmpty(cfg.GroupID, serviceName, "axis")
	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V3_6_0_0
	saramaCfg.ClientID = clientID
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaCfg.Net.DialTimeout = 5 * time.Second
	saramaCfg.Net.ReadTimeout = 5 * time.Second
	saramaCfg.Net.WriteTimeout = 5 * time.Second

	group, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}
	producer, err := newProducer(cfg.Brokers, clientID)
	if err != nil {
		_ = group.Close()
		return nil, err
	}
	ctx, cancel := context.WithCancel(parent)
	dlq := event.AuditTopic + ".dlq"
	g := &Group{cancel: cancel, done: make(chan struct{}), group: group, producer: producer}
	go g.loop(ctx, topics, dlq, handle)
	return g, nil
}

// Close stops the consume loop and closes the group and the dead-letter producer.
func (g *Group) Close() error {
	if g == nil {
		return nil
	}
	g.cancel()
	<-g.done
	err := g.group.Close()
	if closeErr := g.producer.Close(); err == nil {
		err = closeErr
	}
	return err
}

func (g *Group) loop(ctx context.Context, topics []string, dlq string, handle Handler) {
	defer close(g.done)
	session := &groupHandler{producer: g.producer, dlq: dlq, handle: handle}
	for {
		if err := g.group.Consume(ctx, topics, session); err != nil && ctx.Err() == nil {
			logger.WithError(err).Error("kafka consume")
		}
		if ctx.Err() != nil {
			return
		}
	}
}

type groupHandler struct {
	producer *Producer
	dlq      string
	handle   Handler
}

func (h *groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		delivery := Delivery{
			Topic:     msg.Topic,
			Partition: int(msg.Partition),
			Offset:    msg.Offset,
			Key:       msg.Key,
			Value:     msg.Value,
			Headers:   headersOf(msg),
		}
		err := h.handle(session.Context(), delivery)
		if err != nil {
			logger.WithError(err).Error("kafka handler", "topic", msg.Topic, "offset", msg.Offset)
			if dlqErr := h.deadLetter(session.Context(), msg, err); dlqErr != nil {
				return dlqErr
			}
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *groupHandler) deadLetter(ctx context.Context, msg *sarama.ConsumerMessage, cause error) error {
	if msg.Topic == h.dlq {
		return nil
	}
	headers := make([]sarama.RecordHeader, 0, len(msg.Headers)+1)
	for _, header := range msg.Headers {
		if header == nil {
			continue
		}
		headers = append(headers, *header)
	}
	headers = append(headers, sarama.RecordHeader{Key: []byte("error"), Value: []byte(cause.Error())})
	return h.producer.Send(ctx, []*sarama.ProducerMessage{{
		Topic:   h.dlq,
		Key:     sarama.ByteEncoder(msg.Key),
		Value:   sarama.ByteEncoder(msg.Value),
		Headers: headers,
	}})
}

func headersOf(msg *sarama.ConsumerMessage) map[string]string {
	if len(msg.Headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(msg.Headers))
	for _, header := range msg.Headers {
		out[string(header.Key)] = string(header.Value)
	}
	return out
}
