package contextx

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

func FromGinContext(c *gin.Context) *Reader {
	return FromContext(c.Request.Context())
}

func ExtractFromGinHeaders(c *gin.Context) *Builder {
	builder := NewBuilder()

	if reqID := c.GetHeader(HeaderRequestID); reqID != "" {
		builder.WithRequestID(reqID)
	} else if traceID := c.GetHeader(HeaderTraceID); traceID != "" {
		builder.WithRequestID(traceID)
	}

	if parent := c.GetHeader(HeaderParentSpanID); parent != "" {
		builder.WithParentSpanID(parent)
	} else if spanID := c.GetHeader(HeaderSpanID); spanID != "" {
		builder.WithParentSpanID(spanID)
	}

	if auth := c.GetHeader(HeaderAuthorization); auth != "" {
		builder.WithAuthorization(auth)
	}

	clientIP := c.GetHeader(HeaderClientIP)
	if clientIP == "" {
		if forwarded := c.GetHeader(HeaderForwardedFor); forwarded != "" {
			if idx := strings.Index(forwarded, ","); idx > 0 {
				clientIP = strings.TrimSpace(forwarded[:idx])
			} else {
				clientIP = forwarded
			}
		}
	}
	if clientIP == "" {
		clientIP = c.ClientIP()
	}
	builder.WithClientIP(clientIP)

	if ua := c.GetHeader(HeaderUserAgent); ua != "" {
		builder.WithUserAgent(ua)
	}

	return builder
}

func InjectIntoGinHeaders(c *gin.Context, ctx context.Context) {
	reader := FromContext(ctx)

	if reqID := reader.RequestID(); reqID != "" {
		c.Header(HeaderRequestID, reqID)
		c.Header(HeaderTraceID, reqID)
	}
	if spanID := reader.SpanID(); spanID != "" {
		c.Header(HeaderSpanID, spanID)
	}
}
