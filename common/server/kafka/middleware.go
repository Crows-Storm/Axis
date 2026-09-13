package kafka

import (
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/IBM/sarama"
)

type ProducerMiddleware func(msg *sarama.ProducerMessage, evt event.DomainEvent) error

// TracingMiddleware Tracing middleware — Injecting trace_id into Kafka headers
func TracingMiddleware() ProducerMiddleware {
	return func(msg *sarama.ProducerMessage, evt event.DomainEvent) error {
		meta := evt.EventMetadata()
		if meta.TraceId != "" {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   []byte("trace_id"),
				Value: []byte(meta.TraceId),
			})
		}
		// TODO: OpenTelemetry is not yet integrated
		if meta.SpanId != "" {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   []byte("span_id"),
				Value: []byte(meta.SpanId),
			})
		}
		return nil
	}
}

// CorrelationIDMiddleware Association ID Middleware
func CorrelationIDMiddleware() ProducerMiddleware {
	return func(msg *sarama.ProducerMessage, evt event.DomainEvent) error {
		if evt.EventMetadata().CorrID != "" {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   []byte("correlation_id"),
				Value: []byte(evt.EventMetadata().CorrID),
			})
		}
		return nil
	}
}

// AuditMiddleware Auditing Middleware — Add Publisher Information
func AuditMiddleware(serviceName string) ProducerMiddleware {
	return func(msg *sarama.ProducerMessage, evt event.DomainEvent) error {
		msg.Headers = append(msg.Headers,
			sarama.RecordHeader{Key: []byte("source_service"), Value: []byte(serviceName)},
		)
		return nil
	}
}
