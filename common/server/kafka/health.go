package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// HealthChecker Kafka 健康检查
type HealthChecker struct {
	brokers []string
	timeout time.Duration
}

func NewHealthChecker(brokers []string) *HealthChecker {
	return &HealthChecker{
		brokers: brokers,
		timeout: 5 * time.Second,
	}
}

func (h *HealthChecker) Check(ctx context.Context) error {
	cfg := sarama.NewConfig()
	cfg.Net.DialTimeout = h.timeout
	cfg.Net.ReadTimeout = h.timeout
	cfg.Net.WriteTimeout = h.timeout
	cfg.Version = sarama.V3_0_0_0

	client, err := sarama.NewClient(h.brokers, cfg)
	if err != nil {
		return fmt.Errorf("kafka health check failed: %w", err)
	}
	defer client.Close()

	// 验证能获取 cluster 信息
	if len(client.Brokers()) == 0 {
		return fmt.Errorf("kafka health check: no brokers available")
	}

	return nil
}
