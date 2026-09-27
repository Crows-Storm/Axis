package contextx

type (
	requestIDKey     struct{}
	correlationIDKey struct{}
	spanIDKey        struct{}
	parentSpanIDKey  struct{}
	instanceIDKey    struct{}
	authorizationKey struct{}
	clientIPKey      struct{}
	userAgentKey     struct{}
	timestampKey     struct{}
)

var (
	RequestIDKey     = requestIDKey{}
	CorrelationIDKey = correlationIDKey{}
	SpanIDKey        = spanIDKey{}
	ParentSpanIDKey  = parentSpanIDKey{}
	InstanceIDKey    = instanceIDKey{}
	AuthorizationKey = authorizationKey{}
	ClientIPKey      = clientIPKey{}
	UserAgentKey     = userAgentKey{}
	TimestampKey     = timestampKey{}
)
