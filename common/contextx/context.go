package contextx

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/util"
)

type scopeKey struct{}

// Scope is the request-scoped trace bag. One value on context.Context.
// Principal is stored separately via principal.WithPrincipal and is never projected to headers.
type Scope struct {
	instanceID string

	requestID     string
	authorization string
	clientIP      string
	userAgent     string

	correlationID string
	causationId   string
	spanID        string
	parentSpanID  string

	timestamp time.Time
}

type Builder struct {
	scope     *Scope
	principal *principal.Principal
}

func NewBuilder() *Builder {
	return &Builder{
		scope: &Scope{
			timestamp: time.Now(),
		},
	}
}

func (b *Builder) WithRequestID(id string) *Builder {
	b.scope.requestID = id
	return b
}

func (b *Builder) WithCorrelationID(id string) *Builder {
	b.scope.correlationID = id
	return b
}

func (b *Builder) WithSpanID(id string) *Builder {
	b.scope.spanID = id
	return b
}

func (b *Builder) WithParentSpanID(id string) *Builder {
	b.scope.parentSpanID = id
	return b
}

func (b *Builder) WithInstanceID(id string) *Builder {
	b.scope.instanceID = id
	return b
}

func (b *Builder) WithAuthorization(token string) *Builder {
	b.scope.authorization = token
	return b
}

func (b *Builder) WithPrincipal(p *principal.Principal) *Builder {
	b.principal = p
	return b
}

func (b *Builder) WithClientIP(ip string) *Builder {
	b.scope.clientIP = ip
	return b
}

func (b *Builder) WithUserAgent(ua string) *Builder {
	b.scope.userAgent = ua
	return b
}

func (b *Builder) HasRequestID() bool {
	return b.scope.requestID != ""
}

func (b *Builder) Build(ctx context.Context) context.Context {
	// CorrelationID stays with one business operation. The edge generates it
	// when no upstream hop supplied one. CausationID names this hop only and
	// is never copied from the inbound request.
	if b.scope.correlationID == "" {
		b.scope.correlationID = util.UUIDV4().String()
	}
	if b.scope.causationId == "" {
		b.scope.causationId = util.UUIDV4().String()
	}
	if b.scope.spanID == "" {
		b.scope.spanID = generateSpanID()
	}

	ctx = context.WithValue(ctx, scopeKey{}, b.scope)
	if b.principal != nil {
		ctx = principal.WithPrincipal(ctx, b.principal)
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
	return r.scope().requestID
}

func (r *Reader) CorrelationID() string {
	return r.scope().correlationID
}

func (r *Reader) CausationID() string {
	return r.scope().causationId
}

func (r *Reader) SpanID() string {
	return r.scope().spanID
}

func (r *Reader) ParentSpanID() string {
	return r.scope().parentSpanID
}

func (r *Reader) InstanceID() string {
	return r.scope().instanceID
}

func (r *Reader) Authorization() string {
	return r.scope().authorization
}

func (r *Reader) Principal() *principal.Principal {
	return principal.FromContext(r.ctx)
}

func (r *Reader) UserID() uint64 {
	if p := r.Principal(); p != nil {
		return p.UserID
	}
	return 0
}

func (r *Reader) IsAuthenticated() bool {
	return r.Principal() != nil && r.UserID() > 0
}

func (r *Reader) ClientIP() string {
	return r.scope().clientIP
}

func (r *Reader) UserAgent() string {
	return r.scope().userAgent
}

func (r *Reader) Timestamp() time.Time {
	return r.scope().timestamp
}

func (r *Reader) Context() context.Context {
	return r.ctx
}

func (r *Reader) scope() *Scope {
	if r == nil || r.ctx == nil {
		return &Scope{}
	}
	if s, ok := r.ctx.Value(scopeKey{}).(*Scope); ok && s != nil {
		return s
	}
	return &Scope{}
}

func generateSpanID() string {
	return util.UUIDV4().String()[:8]
}
