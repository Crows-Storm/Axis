package kafka

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics Prometheus metrics for Kafka operations
type Metrics struct {
	eventsPublished prometheus.Counter
	eventsConsumed  prometheus.Counter
	eventsFailed    prometheus.Counter
	processingTime  prometheus.Histogram
	deadLetterCount prometheus.Counter
	consumerLag     prometheus.Gauge
}

func NewMetrics(groupOrService string) *Metrics {
	labels := prometheus.Labels{"group": groupOrService}

	return &Metrics{
		eventsPublished: promauto.NewCounter(prometheus.CounterOpts{
			Namespace:   "kafka",
			Subsystem:   "producer",
			Name:        "events_published_total",
			Help:        "Total number of events published",
			ConstLabels: labels,
		}),
		eventsConsumed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace:   "kafka",
			Subsystem:   "consumer",
			Name:        "events_consumed_total",
			Help:        "Total number of events consumed",
			ConstLabels: labels,
		}),
		eventsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace:   "kafka",
			Subsystem:   "consumer",
			Name:        "events_failed_total",
			Help:        "Total number of events failed processing",
			ConstLabels: labels,
		}),
		processingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "kafka",
			Subsystem:   "consumer",
			Name:        "event_processing_duration_seconds",
			Help:        "Event processing duration",
			Buckets:     prometheus.DefBuckets,
			ConstLabels: labels,
		}),
		deadLetterCount: promauto.NewCounter(prometheus.CounterOpts{
			Namespace:   "kafka",
			Subsystem:   "consumer",
			Name:        "dead_letter_total",
			Help:        "Total messages sent to dead letter queue",
			ConstLabels: labels,
		}),
		consumerLag: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   "kafka",
			Subsystem:   "consumer",
			Name:        "consumer_lag",
			Help:        "Consumer group lag",
			ConstLabels: labels,
		}),
	}
}

func (m *Metrics) RecordProcessing(eventName string, duration time.Duration, failed bool) {
	m.eventsConsumed.Inc()
	m.processingTime.Observe(duration.Seconds())
	if failed {
		m.eventsFailed.Inc()
	}
}
