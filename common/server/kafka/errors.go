package kafka

import "errors"

var (
	ErrProducerClosed    = errors.New("kafka: producer is closed")
	ErrConsumerClosed    = errors.New("kafka: consumer is closed")
	ErrSerialization     = errors.New("kafka: serialization failed")
	ErrDeserialization   = errors.New("kafka: deserialization failed")
	ErrNoHandlerFound    = errors.New("kafka: no handler found for event")
	ErrProcessingFailed  = errors.New("kafka: event processing failed")
	ErrBrokerUnavailable = errors.New("kafka: broker unavailable")
	ErrTopicNotFound     = errors.New("kafka: topic not found")
)
