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
)

// MQPropagationKeys is the allow-list projected onto Kafka headers / Envelope.Metadata.
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
