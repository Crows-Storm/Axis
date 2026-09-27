package contextx

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/util"
)

type RequestContext struct {
	requestID     string
	correlationID string
	spanID        string
	parentSpanID  string
	instanceID    string
	authorization string
	principal     *principal.Principal
	clientIP      string
	userAgent     string
	timestamp     time.Time
}

type Builder struct {
	rc *RequestContext
}

func NewBuilder() *Builder {
	return &Builder{
		rc: &RequestContext{
			timestamp: time.Now(),
		},
	}
}

func (b *Builder) WithRequestID(id string) *Builder {
	b.rc.requestID = id
	return b
}

func (b *Builder) WithCorrelationID(id string) *Builder {
	b.rc.correlationID = id
	return b
}

func (b *Builder) WithSpanID(id string) *Builder {
	b.rc.spanID = id
	return b
}

func (b *Builder) WithParentSpanID(id string) *Builder {
	b.rc.parentSpanID = id
	return b
}

func (b *Builder) WithInstanceID(id string) *Builder {
	b.rc.instanceID = id
	return b
}

func (b *Builder) WithAuthorization(token string) *Builder {
	b.rc.authorization = token
	return b
}

func (b *Builder) WithPrincipal(p *principal.Principal) *Builder {
	b.rc.principal = p
	return b
}

func (b *Builder) WithClientIP(ip string) *Builder {
	b.rc.clientIP = ip
	return b
}

func (b *Builder) WithUserAgent(ua string) *Builder {
	b.rc.userAgent = ua
	return b
}

func (b *Builder) HasRequestID() bool {
	return b.rc.requestID != ""
}

func (b *Builder) Build(ctx context.Context) context.Context {
	if b.rc.correlationID == "" {
		b.rc.correlationID = util.UUIDV4().String()
	}
	if b.rc.spanID == "" {
		b.rc.spanID = generateSpanID()
	}

	ctx = context.WithValue(ctx, RequestIDKey, b.rc.requestID)
	ctx = context.WithValue(ctx, CorrelationIDKey, b.rc.correlationID)
	ctx = context.WithValue(ctx, SpanIDKey, b.rc.spanID)
	ctx = context.WithValue(ctx, ParentSpanIDKey, b.rc.parentSpanID)
	ctx = context.WithValue(ctx, InstanceIDKey, b.rc.instanceID)
	ctx = context.WithValue(ctx, ClientIPKey, b.rc.clientIP)
	ctx = context.WithValue(ctx, UserAgentKey, b.rc.userAgent)
	ctx = context.WithValue(ctx, TimestampKey, b.rc.timestamp)

	if b.rc.authorization != "" {
		ctx = WithAuthorization(ctx, b.rc.authorization)
	}
	if b.rc.principal != nil {
		ctx = principal.WithPrincipal(ctx, b.rc.principal)
	}

	return ctx
}

type Reader struct {
	ctx context.Context
}

func FromContext(ctx context.Context) *Reader {
	return &Reader{ctx: ctx}
}

func (r *Reader) RequestID() string {
	return stringValue(r.ctx, RequestIDKey)
}

func (r *Reader) CorrelationID() string {
	return stringValue(r.ctx, CorrelationIDKey)
}

func (r *Reader) SpanID() string {
	return stringValue(r.ctx, SpanIDKey)
}

func (r *Reader) ParentSpanID() string {
	return stringValue(r.ctx, ParentSpanIDKey)
}

func (r *Reader) InstanceID() string {
	return stringValue(r.ctx, InstanceIDKey)
}

func (r *Reader) Authorization() string {
	return stringValue(r.ctx, AuthorizationKey)
}

func (r *Reader) Principal() *principal.Principal {
	return principal.FromContext(r.ctx)
}

func (r *Reader) UserID() int64 {
	if p := r.Principal(); p != nil {
		return p.UserId
	}
	return 0
}

func (r *Reader) IsAuthenticated() bool {
	return r.Principal() != nil && r.UserID() > 0
}

func (r *Reader) ClientIP() string {
	return stringValue(r.ctx, ClientIPKey)
}

func (r *Reader) UserAgent() string {
	return stringValue(r.ctx, UserAgentKey)
}

func (r *Reader) Timestamp() time.Time {
	if ts, ok := r.ctx.Value(TimestampKey).(time.Time); ok {
		return ts
	}
	return time.Time{}
}

func (r *Reader) Context() context.Context {
	return r.ctx
}

func WithAuthorization(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, AuthorizationKey, token)
}

func stringValue(ctx context.Context, key any) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

func generateSpanID() string {
	return util.UUIDV4().String()[:8]
}
