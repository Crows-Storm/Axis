package contextx

import (
	"context"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"google.golang.org/grpc"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = InjectIntoGRPCMetadata(ctx)
		reader := FromContext(ctx)
		logger.Info("gRPC outbound",
			"method", method,
			"request_id", reader.RequestID(),
			"correlation_id", reader.CorrelationID(),
			"causation_id", reader.CausationID(),
			"span_id", reader.SpanID(),
		)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = InjectIntoGRPCMetadata(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
