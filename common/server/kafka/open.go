package kafka

import (
	"fmt"
	"strings"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/domain/event"
)

// OpenPublisher connects a producer and returns an EventBus.
// mirrorAudit copies each business record onto event.AuditTopic.
func OpenPublisher(cfg config.KafkaConfig, serviceName string, mirrorAudit bool) (decorator.EventBus, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("kafka brokers are empty")
	}
	clientID := firstNonEmpty(cfg.ClientID, serviceName, "axis")
	producer, err := newProducer(cfg.Brokers, clientID)
	if err != nil {
		return nil, err
	}
	auditTopic := ""
	if mirrorAudit {
		auditTopic = event.AuditTopic
	}
	return decorator.NewBus(&transport{producer: producer}, serviceName, auditTopic), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
