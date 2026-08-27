package server

import (
	"context"
	"net"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/discovery/registry"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	grpc_logger "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// TokenParser parses a bearer access token into an authenticated principal.
// *jwt.JWTIssuer satisfies this interface.
type TokenParser interface {
	Parse(ctx context.Context, accessToken string) (*principal.Principal, error)
	Preprocessing(ctx context.Context, token string) (string, error)
}

// GRPCServerOption configures optional behavior of RunGRPCServerWithLifecycle.
type GRPCServerOption func(*grpcServerConfig)

type grpcServerConfig struct {
	// tokenParser, when set, lets the inbound interceptor parse the propagated
	// "authorization" metadata and restore the principal into the handler context.
	tokenParser TokenParser
}

// WithGRPCAuthParser enables inbound principal restoration on the gRPC server:
// the "authorization" metadata propagated by upstream services is parsed with
// this service's own signing key and the resulting principal is placed in the
// handler context (grpcx.PrincipalFromContext).
func WithGRPCAuthParser(p TokenParser) GRPCServerOption {
	return func(c *grpcServerConfig) { c.tokenParser = p }
}

// RunGRPCServerWithLifecycle starts the gRPC server and supports graceful shutdown
// ctx: The context used to receive shutdown signals
// registerServer: Callback functions used to register the gRPC service
// registrar: Consul service registrar
// opts: optional server configuration (see WithGRPCAuthParser)
// Returns an error channel, allowing the caller to listen for startup errors
func RunGRPCServerWithLifecycle(
	ctx context.Context,
	registerServer func(server *grpc.Server),
	registrar *registry.Registrar,
	opts ...GRPCServerOption,
) chan error {
	cfg := grpcServerConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	addr := registrar.GetAddress()
	serviceName := registrar.GetName()

	loggerEntry := logger.Entry()
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			inboundMetadataInterceptor(cfg.tokenParser),
			grpc_tags.UnaryServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logger.UnaryServerInterceptor(loggerEntry),
		),
		grpc.ChainStreamInterceptor(
			inboundStreamMetadataInterceptor(cfg.tokenParser),
			grpc_tags.StreamServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logger.StreamServerInterceptor(loggerEntry),
		),
	)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	registerServer(grpcServer)

	if err := registrar.Register(ctx); err != nil {
		logger.Error("consul register failed", "error", err)
		errCh := make(chan error, 1)
		errCh <- err
		close(errCh)
		return errCh
	}

	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("listen failed", "error", err)
		errCh := make(chan error, 1)
		errCh <- err
		close(errCh)
		return errCh
	}
	logger.Info("gRPC server started", "addr", addr)

	errCh := make(chan error, 1)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
			errCh <- err
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := registrar.HealthCheck(); err != nil {
					logger.Error("consul ttl heartbeat failed", "error", err)
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Info("gRPC server shutting down...")

		// set service status to NOT_SERVING
		healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		time.Sleep(3 * time.Second)

		// deregister service from consul
		if err := registrar.Deregister(); err != nil {
			logger.Error("consul deregister failed", "error", err)
		}

		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
			logger.Info("gRPC server stopped gracefully")
		case <-time.After(30 * time.Second):
			logger.Warn("graceful shutdown timeout, forcing stop")
			grpcServer.Stop()
		}

		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		if err := grpcx.CloseAll(closeCtx); err != nil {
			logger.Error("failed to close gRPC client connections", "error", err)
		}

		close(errCh)
	}()

	return errCh
}

// inboundMetadataInterceptor restores request-scoped values propagated by upstream
// services (via grpcx.metadataPropagationInterceptor) back into the handler context:
//
//   - "x-request-id" / "x-trace-id" → grpcx.WithRequestID
//     (business code reads it with grpcx.RequestIDFromContext)
//   - "authorization" → parsed with THIS service's own key via tokenParser,
//     the resulting principal is restored with grpcx.WithPrincipal
//     (business code reads it with grpcx.PrincipalFromContext)
//
// Restoration is best-effort: a missing or invalid token never rejects the call,
// because internal calls (e.g. the login flow) legitimately carry no token.
// Services that require an authenticated principal must check
// grpcx.PrincipalFromContext(ctx) != nil themselves.
func inboundMetadataInterceptor(parser TokenParser) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(restoreInboundContext(ctx, parser), req)
	}
}

func inboundStreamMetadataInterceptor(parser TokenParser) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &inboundStream{ServerStream: ss, ctx: restoreInboundContext(ss.Context(), parser)})
	}
}

// inboundStream overrides Context so stream handlers observe the restored context.
type inboundStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *inboundStream) Context() context.Context { return s.ctx }

func restoreInboundContext(ctx context.Context, parser TokenParser) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	// Request ID: prefer x-request-id, fall back to x-trace-id.
	if vals := md.Get(grpcx.KeyRequestID); len(vals) > 0 && vals[0] != "" {
		ctx = grpcx.WithRequestID(ctx, vals[0])
	} else if vals := md.Get(grpcx.KeyTraceID); len(vals) > 0 && vals[0] != "" {
		ctx = grpcx.WithRequestID(ctx, vals[0])
	}

	// Principal: parse the propagated token with this service's own key (zero-trust:
	// never trust a serialized principal from the wire).
	if parser != nil {
		if vals := md.Get(grpcx.KeyAuthorization); len(vals) > 0 && vals[0] != "" {
			// token preprocessing
			token := vals[0]
			token, err := parser.Preprocessing(ctx, token)
			if err != nil {
				return ctx
			}
			if p, err := parser.Parse(ctx, token); err == nil {
				ctx = grpcx.WithPrincipal(ctx, p)
			}
		}
	}

	return ctx
}
