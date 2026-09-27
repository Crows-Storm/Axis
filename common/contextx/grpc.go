package contextx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func ExtractFromGRPCMetadata(ctx context.Context) *Builder {
	builder := NewBuilder()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return builder
	}

	if values := md.Get(MetadataRequestID); len(values) > 0 && values[0] != "" {
		builder.WithRequestID(values[0])
	} else if values := md.Get(MetadataTraceID); len(values) > 0 && values[0] != "" {
		builder.WithRequestID(values[0])
	}

	if values := md.Get(MetadataCorrelationID); len(values) > 0 && values[0] != "" {
		builder.WithCorrelationID(values[0])
	}

	if values := md.Get(MetadataParentSpanID); len(values) > 0 && values[0] != "" {
		builder.WithParentSpanID(values[0])
	} else if values := md.Get(MetadataSpanID); len(values) > 0 && values[0] != "" {
		builder.WithParentSpanID(values[0])
	}

	if values := md.Get(MetadataAuthorization); len(values) > 0 && values[0] != "" {
		builder.WithAuthorization(values[0])
	}

	if values := md.Get(MetadataClientIP); len(values) > 0 && values[0] != "" {
		builder.WithClientIP(values[0])
	}

	if values := md.Get(MetadataUserAgent); len(values) > 0 && values[0] != "" {
		builder.WithUserAgent(values[0])
	}

	return builder
}

func InjectIntoGRPCMetadata(ctx context.Context) context.Context {
	reader := FromContext(ctx)
	md := metadata.MD{}

	if reqID := reader.RequestID(); reqID != "" {
		md.Set(MetadataRequestID, reqID)
		md.Set(MetadataTraceID, reqID)
	}
	if corrID := reader.CorrelationID(); corrID != "" {
		md.Set(MetadataCorrelationID, corrID)
	}
	if spanID := reader.SpanID(); spanID != "" {
		md.Set(MetadataParentSpanID, spanID)
	}
	if auth := reader.Authorization(); auth != "" {
		md.Set(MetadataAuthorization, auth)
	}
	if clientIP := reader.ClientIP(); clientIP != "" {
		md.Set(MetadataClientIP, clientIP)
	}
	if userAgent := reader.UserAgent(); userAgent != "" {
		md.Set(MetadataUserAgent, userAgent)
	}

	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(md, existingMD)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
