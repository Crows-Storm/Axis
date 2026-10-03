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
	bindTrace(builder, func(key string) string {
		values := md.Get(key)
		if len(values) > 0 {
			return values[0]
		}
		return ""
	}, true)
	return builder
}

func InjectIntoGRPCMetadata(ctx context.Context) context.Context {
	md := metadata.MD{}
	injectTrace(FromContext(ctx), func(key, value string) {
		if value != "" {
			md.Set(key, value)
		}
	}, true)

	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(md, existingMD)
	}
	return metadata.NewOutgoingContext(ctx, md)
}
