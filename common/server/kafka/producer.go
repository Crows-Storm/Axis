package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	asyncProd    sarama.AsyncProducer
	config       ProducerConfig
	mu           sync.RWMutex
	closed       bool
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.ClientID = cfg.ClientID
	saramaCfg.Version = sarama.V3_0_0_0

	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true
	saramaCfg.Producer.RequiredAcks = sarama.RequiredAcks(cfg.ACKs)
	saramaCfg.Producer.Retry.Max = cfg.MaxRetries
	saramaCfg.Producer.Retry.Backoff = cfg.RetryBackoff
	saramaCfg.Producer.MaxMessageBytes = cfg.MaxMessageBytes

	switch cfg.Compression {
	case "gzip":
		saramaCfg.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		saramaCfg.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		saramaCfg.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		saramaCfg.Producer.Compression = sarama.CompressionZSTD
	}

	if cfg.Idempotent {
		saramaCfg.Producer.Idempotent = true
		saramaCfg.Net.MaxOpenRequests = 1
	}

	// Batch configuration
	saramaCfg.Producer.Flush.Bytes = cfg.BatchSize
	saramaCfg.Producer.Flush.Frequency = cfg.BatchTimeout

	// TLS / SASL
	applySecurity(saramaCfg, cfg.TLSEnabled, cfg.SASLMechanism, cfg.SASLUsername, cfg.SASLPassword)

	syncProd, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka sync producer: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"brokers":    cfg.Brokers,
		"client_id":  cfg.ClientID,
		"idempotent": cfg.Idempotent,
	})

	return &Producer{
		syncProducer: syncProd,
		config:       cfg,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) (int32, int64, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return 0, 0, ErrProducerClosed
	}
	p.mu.RUnlock()

	done := make(chan sendResult, 1)
	go func() {
		partition, offset, err := p.syncProducer.SendMessage(msg)
		done <- sendResult{partition, offset, err}
	}()

	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	case result := <-done:
		return result.partition, result.offset, result.err
	}
}

func (p *Producer) SendMessages(ctx context.Context, msgs []*sarama.ProducerMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return ErrProducerClosed
	}
	p.mu.RUnlock()

	done := make(chan error, 1)
	go func() {
		done <- p.syncProducer.SendMessages(msgs)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	logger.Info("Closing Kafka producer...")
	return p.syncProducer.Close()
}

type sendResult struct {
	partition int32
	offset    int64
	err       error
}

// applySecurity configuration TLS/SASL
func applySecurity(cfg *sarama.Config, tlsEnabled bool, mechanism, username, password string) {
	if tlsEnabled {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	if mechanism != "" {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = username
		cfg.Net.SASL.Password = password

		switch mechanism {
		case "SCRAM-SHA-256":
			cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
			}
		case "SCRAM-SHA-512":
			cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
			}
		default:
			cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		}
	}
}
