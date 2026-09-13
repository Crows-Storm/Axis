package outbox

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// Processor 后台轮询 Outbox 并发布到 Kafka
type Processor struct {
	repo      OutboxRepository
	publisher event.EventBus
	logger    zerolog.Logger
	interval  time.Duration
	batchSize int
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewProcessor(
	repo OutboxRepository,
	publisher event.EventBus,
	interval time.Duration,
	batchSize int,
	logger zerolog.Logger,
) *Processor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Processor{
		repo:      repo,
		publisher: publisher,
		logger:    logger.With().Str("component", "outbox-processor").Logger(),
		interval:  interval,
		batchSize: batchSize,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动 outbox 处理器
func (p *Processor) Start() {
	p.logger.Info().
		Dur("interval", p.interval).
		Int("batch_size", p.batchSize).
		Msg("Starting outbox processor")

	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-p.ctx.Done():
				p.logger.Info().Msg("Outbox processor stopped")
				return
			case <-ticker.C:
				p.processBatch()
			}
		}
	}()
}

func (p *Processor) processBatch() {
	msgs, err := p.repo.FindPending(p.ctx, p.batchSize)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to fetch pending outbox messages")
		return
	}

	if len(msgs) == 0 {
		return
	}

	p.logger.Debug().Int("count", len(msgs)).Msg("Processing outbox batch")

	for _, msg := range msgs {
		// 构建领域事件 envelope 并发布
		evt := &outboxEvent{
			name:        msg.EventName,
			aggregateID: msg.AggregateID,
			payload:     msg.Payload,
			metadata:    msg.Metadata,
		}

		if err := p.publisher.Publish(p.ctx, evt); err != nil {
			p.logger.Error().Err(err).
				Str("outbox_id", msg.ID).
				Str("event", msg.EventName).
				Msg("Failed to publish outbox message")

			_ = p.repo.MarkFailed(p.ctx, msg.ID, err.Error())
			_ = p.repo.IncrementRetry(p.ctx, msg.ID)
			continue
		}

		_ = p.repo.MarkPublished(p.ctx, msg.ID)
	}
}

// Stop 停止
func (p *Processor) Stop() {
	p.cancel()
}

// outboxEvent 内部适配类型
type outboxEvent struct {
	name        string
	aggregateID string
	payload     string
	metadata    string
}

func (e *outboxEvent) EventName() string     { return e.name }
func (e *outboxEvent) AggregateID() string   { return e.aggregateID }
func (e *outboxEvent) OccurredAt() time.Time { return time.Now() }
func (e *outboxEvent) EventVersion() int     { return 1 }
func (e *outboxEvent) Payload() map[string]interface{} {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(e.payload), &m)
	return m
}
func (e *outboxEvent) Metadata() map[string]string {
	var m map[string]string
	_ = json.Unmarshal([]byte(e.metadata), &m)
	return m
}
