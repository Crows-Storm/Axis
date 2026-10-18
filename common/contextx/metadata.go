package contextx

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderTraceID       = "X-Trace-ID"
	HeaderSpanID        = "X-Span-ID"
	HeaderParentSpanID  = "X-Parent-Span-ID"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderClientIP      = "X-Real-IP"
	HeaderForwardedFor  = "X-Forwarded-For"
)

const (
	MetadataRequestID     = "x-request-id"
	MetadataCorrelationID = "x-correlation-id"
	MetadataCausationId   = "x-causation-id"
	MetadataTraceID       = "x-trace-id"
	MetadataSpanID        = "x-span-id"
	MetadataParentSpanID  = "x-parent-span-id"
	MetadataAuthorization = "authorization"
	MetadataClientIP      = "x-client-ip"
	MetadataUserAgent     = "x-user-agent"

	// Venue API credentials for internal exchange gRPC. Never log or put on MQ.
	MetadataExchangeAPIKey     = "x-exchange-api-key"
	MetadataExchangeAPISecret  = "x-exchange-api-secret"
	MetadataExchangePassphrase = "x-exchange-passphrase"
)

// MQPropagationKeys is the allow-list projected from request context onto Kafka headers.
// Domain causation (parent event_id) is not in this list from context; it is set on Envelope.Metadata.
// Authorization, instance-id, and principal/user-id are never included.
var MQPropagationKeys = []string{
	MetadataRequestID,
	MetadataCorrelationID,
	MetadataCausationId,
	MetadataTraceID,
	MetadataParentSpanID,
	MetadataClientIP,
	MetadataUserAgent,
}

func IsMQPropagationKey(key string) bool {
	switch key {
	case MetadataRequestID, MetadataCorrelationID, MetadataCausationId, MetadataTraceID, MetadataParentSpanID, MetadataClientIP, MetadataUserAgent:
		return true
	default:
		return false
	}
}
