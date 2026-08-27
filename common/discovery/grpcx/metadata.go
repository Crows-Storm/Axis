package grpcx

import (
	"context"

	principal "github.com/Crows-Storm/Axis/common/domain/principal"
	"google.golang.org/grpc/metadata"
)

// ==================== Metadata Key Constants ====================
// gRPC metadata keys are always lowercase on the wire.

const (
	KeyAuthorization = "authorization"
	KeyRequestID     = "x-request-id"
	KeyTraceID       = "x-trace-id"
)

// allowedPropagationKeys defines which metadata keys are propagated across service boundaries.
// Only explicitly listed keys are forwarded to prevent leaking sensitive headers (e.g. cookies).
var allowedPropagationKeys = map[string]bool{
	KeyAuthorization: true,
	KeyRequestID:     true,
	KeyTraceID:       true,
}

// ==================== Outbound Metadata (HTTP middleware → gRPC client) ====================

// outboundMDKey is the context key for metadata that HTTP middleware attaches
// so the gRPC client interceptor can propagate it to downstream services.
type outboundMDKey struct{}

// WithOutgoingMetadata attaches metadata to context for the gRPC client interceptor to propagate.
// Called by HTTP middleware (e.g. RequestMetadataMiddleware) to bridge HTTP headers → gRPC metadata.
func WithOutgoingMetadata(ctx context.Context, md metadata.MD) context.Context {
	return context.WithValue(ctx, outboundMDKey{}, md)
}

// OutgoingMetadataFromContext retrieves metadata set by HTTP middleware for outbound propagation.
func OutgoingMetadataFromContext(ctx context.Context) (metadata.MD, bool) {
	md, ok := ctx.Value(outboundMDKey{}).(metadata.MD)
	return md, ok
}

// ==================== Inbound Metadata (gRPC server → handler context) ====================

// requestIDKey is the context key for the request ID extracted from incoming gRPC metadata.
type requestIDKey struct{}

// WithRequestID sets the request ID in context (called by inbound interceptor).
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext retrieves the request ID from context.
// Returns empty string if not set.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// WithPrincipal sets the principal in context
func WithPrincipal(ctx context.Context, p *principal.Principal) context.Context {
	return principal.WithPrincipal(ctx, p)
}

// PrincipalFromContext retrieves the principal from context.
// Returns empty string if not set.
func PrincipalFromContext(ctx context.Context) *principal.Principal {
	p := principal.FromContext(ctx)
	return p
}
