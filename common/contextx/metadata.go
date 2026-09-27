package contextx

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
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
	MetadataTraceID       = "x-trace-id"
	MetadataSpanID        = "x-span-id"
	MetadataParentSpanID  = "x-parent-span-id"
	MetadataAuthorization = "authorization"
	MetadataClientIP      = "x-client-ip"
	MetadataUserAgent     = "x-user-agent"
)

var AllowedPropagationKeys = map[string]bool{
	MetadataRequestID:     true,
	MetadataCorrelationID: true,
	MetadataTraceID:       true,
	MetadataParentSpanID:  true,
	MetadataAuthorization: true,
	MetadataClientIP:      true,
	MetadataUserAgent:     true,
}
