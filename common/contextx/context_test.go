package contextx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestBuilder_Build_DoesNotAutoGenerateRequestID(t *testing.T) {
	ctx := NewBuilder().Build(context.Background())
	reader := FromContext(ctx)

	assert.Empty(t, reader.RequestID())
	assert.NotEmpty(t, reader.CorrelationID())
	assert.NotEmpty(t, reader.CausationID())
	assert.NotEqual(t, reader.CorrelationID(), reader.CausationID())
	assert.NotEmpty(t, reader.SpanID())
	assert.Len(t, reader.SpanID(), 8)
	assert.Nil(t, reader.Principal())
	assert.False(t, reader.IsAuthenticated())
}

func TestBuilder_Build_CustomValues(t *testing.T) {
	p := &principal.Principal{UserID: 42}

	ctx := NewBuilder().
		WithRequestID("req-123").
		WithCorrelationID("corr-456").
		WithSpanID("span-789").
		WithParentSpanID("parent-001").
		WithInstanceID("user-127.0.0.1-9000").
		WithAuthorization("Bearer token").
		WithPrincipal(p).
		WithClientIP("192.168.1.1").
		WithUserAgent("TestAgent/1.0").
		Build(context.Background())

	reader := FromContext(ctx)

	assert.Equal(t, "req-123", reader.RequestID())
	assert.Equal(t, "corr-456", reader.CorrelationID())
	assert.NotEmpty(t, reader.CausationID())
	assert.Equal(t, "span-789", reader.SpanID())
	assert.Equal(t, "parent-001", reader.ParentSpanID())
	assert.Equal(t, "user-127.0.0.1-9000", reader.InstanceID())
	assert.Equal(t, "Bearer token", reader.Authorization())
	assert.Equal(t, "192.168.1.1", reader.ClientIP())
	assert.Equal(t, "TestAgent/1.0", reader.UserAgent())
	require.NotNil(t, reader.Principal())
	assert.Equal(t, uint64(42), reader.UserID())
	assert.True(t, reader.IsAuthenticated())
}

func TestExtractFromGinHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(HeaderRequestID, "req-789")
	c.Request.Header.Set("X-Correlation-ID", "forged-corr")
	c.Request.Header.Set("X-Causation-ID", "forged-cause")
	c.Request.Header.Set(HeaderSpanID, "upstream-span")
	c.Request.Header.Set(HeaderAuthorization, "Bearer abc")
	c.Request.Header.Set(HeaderClientIP, "10.0.0.1")
	c.Request.Header.Set("X-Instance-ID", "forged-instance")

	ctx := ExtractFromGinHeaders(c).WithInstanceID("local-instance").Build(c.Request.Context())
	reader := FromContext(ctx)

	assert.Equal(t, "req-789", reader.RequestID())
	assert.NotEmpty(t, reader.CorrelationID())
	assert.NotEqual(t, "forged-corr", reader.CorrelationID())
	assert.NotEmpty(t, reader.CausationID())
	assert.NotEqual(t, "forged-cause", reader.CausationID())
	assert.Equal(t, "upstream-span", reader.ParentSpanID())
	assert.NotEqual(t, "upstream-span", reader.SpanID())
	assert.Equal(t, "local-instance", reader.InstanceID())
	assert.Equal(t, "Bearer abc", reader.Authorization())
	assert.Equal(t, "10.0.0.1", reader.ClientIP())
}

func TestInjectIntoGRPCMetadata_OmitsUserAndInstance(t *testing.T) {
	ctx := NewBuilder().
		WithRequestID("inject-req").
		WithCorrelationID("inject-corr").
		WithSpanID("inject-span").
		WithInstanceID("should-not-leave-process").
		WithAuthorization("Bearer xyz").
		WithPrincipal(&principal.Principal{UserID: 99}).
		Build(context.Background())

	outCtx := InjectIntoGRPCMetadata(ctx)
	md, ok := metadata.FromOutgoingContext(outCtx)
	require.True(t, ok)

	assert.Equal(t, []string{"inject-req"}, md.Get(MetadataRequestID))
	assert.Equal(t, []string{"inject-req"}, md.Get(MetadataTraceID))
	assert.Equal(t, []string{"inject-corr"}, md.Get(MetadataCorrelationID))
	assert.Equal(t, []string{FromContext(ctx).CausationID()}, md.Get(MetadataCausationId))
	assert.Equal(t, []string{"inject-span"}, md.Get(MetadataParentSpanID))
	assert.Equal(t, []string{"Bearer xyz"}, md.Get(MetadataAuthorization))
	assert.Empty(t, md.Get("x-user-id"))
	assert.Empty(t, md.Get("x-instance-id"))
	assert.Empty(t, md.Get(MetadataSpanID))
}

func TestGRPCRoundTrip_KeepsRequestAndCorr_NewSpan(t *testing.T) {
	originalCtx := NewBuilder().
		WithRequestID("roundtrip-req").
		WithCorrelationID("roundtrip-corr").
		WithInstanceID("svc-a").
		WithAuthorization("Bearer rt").
		Build(context.Background())
	upstreamSpan := FromContext(originalCtx).SpanID()
	upstreamCause := FromContext(originalCtx).CausationID()

	outCtx := InjectIntoGRPCMetadata(originalCtx)
	md, ok := metadata.FromOutgoingContext(outCtx)
	require.True(t, ok)

	incomingCtx := metadata.NewIncomingContext(context.Background(), md)
	restored := ExtractFromGRPCMetadata(incomingCtx).
		WithInstanceID("svc-b").
		Build(incomingCtx)
	reader := FromContext(restored)

	assert.Equal(t, "roundtrip-req", reader.RequestID())
	assert.Equal(t, "roundtrip-corr", reader.CorrelationID())
	assert.NotEmpty(t, reader.CausationID())
	assert.NotEqual(t, upstreamCause, reader.CausationID())
	assert.Equal(t, "Bearer rt", reader.Authorization())
	assert.Equal(t, upstreamSpan, reader.ParentSpanID())
	assert.NotEqual(t, upstreamSpan, reader.SpanID())
	assert.Equal(t, "svc-b", reader.InstanceID())
	assert.Nil(t, reader.Principal())
}

func TestInjectIntoMQHeaders_OmitsAuthInstanceAndPrincipal(t *testing.T) {
	ctx := NewBuilder().
		WithRequestID("mq-req").
		WithCorrelationID("mq-corr").
		WithSpanID("mq-span").
		WithInstanceID("must-not-leave").
		WithAuthorization("Bearer secret").
		WithPrincipal(&principal.Principal{UserID: 11}).
		WithClientIP("10.1.1.1").
		Build(context.Background())

	headers := InjectIntoMQHeaders(ctx)
	assert.Equal(t, "mq-req", headers[MetadataRequestID])
	assert.Equal(t, "mq-req", headers[MetadataTraceID])
	assert.Equal(t, "mq-corr", headers[MetadataCorrelationID])
	assert.Equal(t, FromContext(ctx).CausationID(), headers[MetadataCausationId])
	assert.Equal(t, "mq-span", headers[MetadataParentSpanID])
	assert.Equal(t, "10.1.1.1", headers[MetadataClientIP])
	assert.Empty(t, headers[MetadataAuthorization])
	assert.Empty(t, headers[MetadataSpanID])
	assert.NotContains(t, headers, "x-user-id")
	assert.NotContains(t, headers, "x-instance-id")
}

func TestMQRoundTrip_UsesEventIDWhenRequestIDMissing(t *testing.T) {
	published := NewBuilder().WithRequestID("pub-req").WithCorrelationID("pub-corr").Build(context.Background())
	headers := InjectIntoMQHeaders(published)
	incomingSpan := headers[MetadataParentSpanID]
	publishedCause := FromContext(published).CausationID()

	restored := BindMQ(context.Background(), headers, "", "svc-consumer")
	reader := FromContext(restored)
	assert.Equal(t, "pub-req", reader.RequestID())
	assert.Equal(t, "pub-corr", reader.CorrelationID())
	assert.NotEqual(t, publishedCause, reader.CausationID())
	assert.NotEmpty(t, reader.CausationID())
	assert.Equal(t, incomingSpan, reader.ParentSpanID())
	assert.NotEqual(t, incomingSpan, reader.SpanID())
	assert.Equal(t, "svc-consumer", reader.InstanceID())
	assert.Empty(t, reader.Authorization())
	assert.Nil(t, reader.Principal())

	fallback := BindMQ(context.Background(), map[string]string{}, "evt-fallback", "svc-consumer")
	assert.Equal(t, "evt-fallback", FromContext(fallback).RequestID())
}
