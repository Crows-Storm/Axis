package kafka

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// EventPublisher 实现 domain/event.EventBus 接口
// 将领域事件通过 Kafka 发布出去
type EventPublisher struct {
	producer    *Producer
	serializer  *Serializer
	topicRouter *TopicRouter
	middleware  []ProducerMiddleware
	logger      *logrus.Logger
}

func NewEventPublisher(
	producer *Producer,
	serializer *Serializer,
	router *TopicRouter,
	logger *logrus.Logger,
) *EventPublisher {
	return &EventPublisher{
		producer:    producer,
		serializer:  serializer,
		topicRouter: router,
		logger:      logger,
	}
}

// Use 添加 Producer 中间件
func (p *EventPublisher) Use(mw ProducerMiddleware) {
	p.middleware = append(p.middleware, mw)
}

// Publish 实现 EventBus 接口，发布一个或多个领域事件
func (p *EventPublisher) Publish(ctx context.Context, events ...event.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]*sarama.ProducerMessage, 0, len(events))

	for _, evt := range events {
		// 1. 序列化
		payload, err := p.serializer.Serialize(evt)
		if err != nil {
			return fmt.Errorf("serialize event %s: %w", evt.EventName(), err)
		}

		// 2. 路由 topic
		topic := p.topicRouter.Resolve(evt)

		// 3. 构建消息 headers
		headers := buildHeaders(evt)

		msg := &sarama.ProducerMessage{
			Topic:   topic,
			Key:     sarama.StringEncoder(strconv.FormatInt(evt.AggregateId(), 10)), // Messages from the same aggregate root enter the same partition
			Value:   sarama.ByteEncoder(payload),
			Headers: headers,
		}

		// 4. 执行中间件链
		for _, mw := range p.middleware {
			if err := mw(msg, evt); err != nil {
				return fmt.Errorf("middleware error for event %s: %w", evt.EventName(), err)
			}
		}

		messages = append(messages, msg)
	}

	// 5. 发送
	if len(messages) == 1 {
		partition, offset, err := p.producer.SendMessage(ctx, messages[0])
		if err != nil {
			return fmt.Errorf("publish event %s: %w", events[0].EventName(), err)
		}
		p.logger.WithFields(logrus.Fields{
			"event":     events[0].EventName(),
			"topic":     messages[0].Topic,
			"partition": partition,
			"offset":    offset,
		}).Debug("Event published")
	} else {
		if err := p.producer.SendMessages(ctx, messages); err != nil {
			return fmt.Errorf("batch publish events: %w", err)
		}
		p.logger.WithField("count", len(messages)).
			Debug("Events batch published")
	}

	return nil
}

func (p *EventPublisher) Close() error {
	return p.producer.Close()
}

// buildHeaders 构建 Kafka 消息头
func buildHeaders(evt event.DomainEvent) []sarama.RecordHeader {
	headers := []sarama.RecordHeader{
		{Key: []byte("event_name"), Value: []byte(evt.EventName())},
		{Key: []byte("event_version"), Value: []byte(fmt.Sprintf("%d", evt.EventVersion()))},
		{Key: []byte("aggregate_id"), Value: []byte(strconv.FormatInt(evt.AggregateId(), 10))},
		{Key: []byte("content_type"), Value: []byte("application/json")},
	}

	for k, v := range evt.Metadata() {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	return headers
}
