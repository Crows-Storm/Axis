package server

import (
	"context"
	"net"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/discovery/registry"
	grpc_logger "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// RunGRPCServerWithLifecycle starts the gRPC server and supports graceful shutdown
// ctx: The context used to receive shutdown signals
// registerServer: Callback functions used to register the gRPC service
// registrar: Consul service registrar
// Returns an error channel, allowing the caller to listen for startup errors
func RunGRPCServerWithLifecycle(
	ctx context.Context,
	registerServer func(server *grpc.Server),
	registrar *registry.Registrar,
) chan error {
	addr := registrar.GetAddress()
	serviceName := registrar.GetName()

	loggerEntry := logger.Entry()
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_tags.UnaryServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logger.UnaryServerInterceptor(loggerEntry),
		),
		grpc.ChainStreamInterceptor(
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
