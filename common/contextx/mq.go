package contextx

import (
	"context"
	"strings"
)

func ExtractFromMQHeaders(headers map[string]string, fallbackRequestID string) *Builder {
	builder := NewBuilder()
	normalized := normalizeHeaderMap(headers)
	bindTrace(builder, func(key string) string {
		return normalized[key]
	}, false)
	if !builder.HasRequestID() && fallbackRequestID != "" {
		builder.WithRequestID(fallbackRequestID)
	}
	return builder
}

func InjectIntoMQHeaders(ctx context.Context) map[string]string {
	out := make(map[string]string, len(MQPropagationKeys)+1)
	injectTrace(FromContext(ctx), func(key, value string) {
		if value == "" || !IsMQPropagationKey(key) {
			return
		}
		out[key] = value
	}, false)
	return out
}

func BindMQ(ctx context.Context, headers map[string]string, fallbackRequestID, instanceID string) context.Context {
	return ExtractFromMQHeaders(headers, fallbackRequestID).
		WithInstanceID(instanceID).
		Build(ctx)
}

func MergeMQHeaders(dst map[string]string, ctx context.Context) map[string]string {
	if dst == nil {
		dst = make(map[string]string)
	}
	for k, v := range InjectIntoMQHeaders(ctx) {
		if v != "" {
			dst[k] = v
		}
	}
	return dst
}

func normalizeHeaderMap(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		out[strings.ToLower(strings.TrimSpace(k))] = v
	}
	return out
}
