package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// Consumer Kafka 消费者封装，支持优雅关闭、重试、死信
type Consumer struct {
	group      sarama.ConsumerGroup
	config     ConsumerConfig
	dispatcher *event.EventDispatcher
	retry      *RetryStrategy
	deadLetter *DeadLetterProducer
	logger     *logrus.Entry
	metrics    *Metrics

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	ready  chan bool
}

func NewConsumer(
	cfg ConsumerConfig,
	dispatcher *event.EventDispatcher,
	producer *Producer, // 用于死信发送
	logger *logrus.Entry,
) (*Consumer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.ClientID = cfg.ClientID
	saramaCfg.Version = sarama.V3_0_0_0
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyCooperativeSticky(),
	}
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = false // 手动提交
	saramaCfg.Consumer.Offsets.Initial = offsetReset(cfg.AutoOffsetReset)
	saramaCfg.Consumer.Group.Session.Timeout = cfg.SessionTimeout
	saramaCfg.Consumer.Group.Heartbeat.Interval = cfg.HeartbeatInterval
	saramaCfg.Consumer.Fetch.Min = int32(cfg.FetchMinBytes)
	saramaCfg.Consumer.Fetch.Max = 1048576 // 1MB
	saramaCfg.Consumer.MaxProcessingTime = 5 * time.Minute

	applySecurity(saramaCfg, cfg.TLSEnabled, cfg.SASLMechanism, cfg.SASLUsername, cfg.SASLPassword)

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var dlp *DeadLetterProducer
	if cfg.DeadLetterTopic != "" && producer != nil {
		dlp = NewDeadLetterProducer(producer, cfg.DeadLetterTopic)
	}

	c := &Consumer{
		group:      group,
		config:     cfg,
		dispatcher: dispatcher,
		retry:      NewRetryStrategy(cfg.MaxRetries, cfg.RetryBaseDelay, cfg.RetryMaxDelay),
		deadLetter: dlp,
		logger:     logger.WithField("consumer_group", cfg.GroupID),
		metrics:    NewMetrics(cfg.GroupID),
		ctx:        ctx,
		cancel:     cancel,
		ready:      make(chan bool),
	}

	// 后台消费错误监听
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for err := range group.Errors() {
			c.logger.WithError(err).Error("Consumer group error")
		}
	}()

	return c, nil
}

// Start 开始消费
func (c *Consumer) Start() error {
	c.logger.WithField("topics", c.config.Topics).Info("Starting consumer...")

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.group.Consume(c.ctx, c.config.Topics, c); err != nil {
				c.logger.WithError(err).Error("Consumer error, restarting...")
			}
			if c.ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready // 等待首次就绪
	return nil
}

// ─── sarama.ConsumerGroupHandler 接口实现 ─────────────────────────

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.WithField("generation", session.GenerationID()).
		Info("Consumer group setup, partition assignment received")
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.WithField("generation", session.GenerationID()).
		Info("Consumer group cleanup, rebalancing")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			c.processMessage(session, msg)
		}
	}
}

// processMessage 处理单条消息，含重试逻辑
func (c *Consumer) processMessage(session sarama.ConsumerGroupSession, msg *sarama.Message) {
	start := time.Now()
	envelope, err := DeserializeEnvelope(msg)
	if err != nil {
		c.logger.WithFields(logger.Fields{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
		}).WithError(err).Error("Failed to deserialize message, sending to dead letter")
		c.sendToDeadLetter(msg, nil, err)
		session.MarkMessage(msg, "")
		return
	}

	envelope.Topic = msg.Topic
	envelope.Partition = int(msg.Partition)
	envelope.Offset = msg.Offset

	// 带重试地执行 handler
	processErr := c.retry.Do(c.ctx, func(ctx context.Context) error {
		return c.dispatcher.Dispatch(ctx, *envelope)
	})

	duration := time.Since(start)
	c.metrics.RecordProcessing(envelope.EventName, duration, processErr != nil)

	if processErr != nil {
		c.logger.WithFields(logger.Fields{
			"event":    envelope.EventName,
			"event_id": envelope.EventID,
			"retries":  c.config.MaxRetries,
			"duration": duration,
		}).WithError(processErr).Error("Event processing failed after all retries")

		c.sendToDeadLetter(msg, envelope, processErr)
	} else {
		c.logger.WithFields(logger.Fields{
			"event":    envelope.EventName,
			"event_id": envelope.EventID,
			"duration": duration,
		}).Debug("Event processed successfully")
	}

	// 无论成功与否都标记已处理（失败已入死信）
	session.MarkMessage(msg, "")
	// 定期提交 offset
	session.Commit()
}

func (c *Consumer) sendToDeadLetter(msg *sarama.Message, envelope *event.Envelope, processErr error) {
	if c.deadLetter == nil {
		c.logger.Warn("No dead letter topic configured, message dropped")
		return
	}
	if err := c.deadLetter.Send(c.ctx, msg, envelope, processErr); err != nil {
		c.logger.WithError(err).Error("Failed to send to dead letter queue")
	}
}

// Stop 优雅关闭
func (c *Consumer) Stop() error {
	c.logger.Info("Stopping consumer...")
	c.cancel()
	c.wg.Wait()
	return c.group.Close()
}

func offsetReset(reset string) int64 {
	if reset == "latest" {
		return sarama.OffsetNewest
	}
	return sarama.OffsetOldest
}
