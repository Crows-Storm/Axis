package server

import (
	"context"
	"net"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/discovery/registry"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/util"
	grpc_logger "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type TokenParser interface {
	Parse(ctx context.Context, accessToken string) (*principal.Principal, error)
	Preprocessing(ctx context.Context, token string) (string, error)
}

type GRPCServerOption func(*grpcServerConfig)

type grpcServerConfig struct {
	tokenParser TokenParser
	instanceID  string
}

func WithGRPCAuthParser(p TokenParser) GRPCServerOption {
	return func(c *grpcServerConfig) { c.tokenParser = p }
}

func WithGRPCInstanceID(id string) GRPCServerOption {
	return func(c *grpcServerConfig) { c.instanceID = id }
}

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
			unaryContextInterceptor(cfg.tokenParser, cfg.instanceID),
			grpc_tags.UnaryServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logger.UnaryServerInterceptor(loggerEntry),
		),
		grpc.ChainStreamInterceptor(
			streamContextInterceptor(cfg.tokenParser, cfg.instanceID),
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

		healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		time.Sleep(3 * time.Second)

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

func unaryContextInterceptor(parser TokenParser, instanceID string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(restoreInboundContext(ctx, parser, instanceID), req)
	}
}

func streamContextInterceptor(parser TokenParser, instanceID string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &inboundStream{ServerStream: ss, ctx: restoreInboundContext(ss.Context(), parser, instanceID)})
	}
}

type inboundStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *inboundStream) Context() context.Context { return s.ctx }

func restoreInboundContext(ctx context.Context, parser TokenParser, instanceID string) context.Context {
	builder := contextx.ExtractFromGRPCMetadata(ctx)
	builder.WithInstanceID(instanceID)
	if !builder.HasRequestID() {
		builder.WithRequestID(util.UUIDV4().String())
		logger.WithField("instance_id", instanceID).Warn("gRPC inbound missing request id, generated locally")
	}
	ctx = builder.Build(ctx)

	reader := contextx.FromContext(ctx)
	rawAuth := reader.Authorization()
	if parser != nil && rawAuth != "" {
		token, err := parser.Preprocessing(ctx, rawAuth)
		if err != nil {
			return ctx
		}
		if p, err := parser.Parse(ctx, token); err == nil {
			ctx = principal.WithPrincipal(ctx, p)
		}
	}

	return ctx
}
