package kafka

import "time"

// kafka config in yaml or .env?
type ProducerConfig struct {
	Brokers         []string      `yaml:"brokers"`
	ClientID        string        `yaml:"client_id"`
	TopicPrefix     string        `yaml:"topic_prefix"` // 如 "order-service"
	ACKs            int           `yaml:"acks"`         // -1=all, 0=none, 1=leader
	MaxRetries      int           `yaml:"max_retries"`
	RetryBackoff    time.Duration `yaml:"retry_backoff"`
	BatchSize       int           `yaml:"batch_size"` // bytes
	BatchTimeout    time.Duration `yaml:"batch_timeout"`
	LingerMs        int           `yaml:"linger_ms"`
	Compression     string        `yaml:"compression"` // gzip, snappy, lz4, zstd
	Idempotent      bool          `yaml:"idempotent"`  // exactly-once producer
	FlushTimeout    time.Duration `yaml:"flush_timeout"`
	MaxMessageBytes int           `yaml:"max_message_bytes"`
	// TLS / SASL
	TLSEnabled    bool   `yaml:"tls_enabled"`
	SASLMechanism string `yaml:"sasl_mechanism"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	SASLUsername  string `yaml:"sasl_username"`
	SASLPassword  string `yaml:"sasl_password"`
}

type ConsumerConfig struct {
	Brokers           []string      `yaml:"brokers"`
	GroupID           string        `yaml:"group_id"`
	ClientID          string        `yaml:"client_id"`
	Topics            []string      `yaml:"topics"`
	AutoOffsetReset   string        `yaml:"auto_offset_reset"` // earliest, latest
	SessionTimeout    time.Duration `yaml:"session_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	MaxPollRecords    int           `yaml:"max_poll_records"`
	FetchMinBytes     int           `yaml:"fetch_min_bytes"`
	FetchMaxWait      time.Duration `yaml:"fetch_max_wait"`
	// retry config
	MaxRetries     int           `yaml:"max_retries"`
	RetryBaseDelay time.Duration `yaml:"retry_base_delay"`
	RetryMaxDelay  time.Duration `yaml:"retry_max_delay"`
	// Dead letter queue
	DeadLetterTopic string `yaml:"dead_letter_topic"`
	// TLS / SASL
	TLSEnabled    bool   `yaml:"tls_enabled"`
	SASLMechanism string `yaml:"sasl_mechanism"`
	SASLUsername  string `yaml:"sasl_username"`
	SASLPassword  string `yaml:"sasl_password"`
}

// DefaultProducerConfig return a default configuration
func DefaultProducerConfig(brokers []string) ProducerConfig {
	return ProducerConfig{
		Brokers:         brokers,
		Acks:            -1,
		MaxRetries:      3,
		RetryBackoff:    100 * time.Millisecond,
		BatchSize:       16384,
		LingerMs:        5,
		Compression:     "snappy",
		Idempotent:      true,
		FlushTimeout:    10 * time.Second,
		MaxMessageBytes: 1048576, // 1MB
	}
}

// DefaultConsumerConfig return a default configuration
func DefaultConsumerConfig(brokers []string, groupID string) ConsumerConfig {
	return ConsumerConfig{
		Brokers:           brokers,
		GroupID:           groupID,
		AutoOffsetReset:   "earliest",
		SessionTimeout:    30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
		MaxPollRecords:    500,
		FetchMinBytes:     1,
		FetchMaxWait:      500 * time.Millisecond,
		MaxRetries:        5,
		RetryBaseDelay:    1 * time.Second,
		RetryMaxDelay:     30 * time.Second,
	}
}
