package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Crows-Storm/Axis/audit/protos"
	"github.com/Crows-Storm/Axis/audit/service"
	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/discovery/consulx"
	"github.com/Crows-Storm/Axis/common/discovery/registry"
	"github.com/Crows-Storm/Axis/common/genproto/auditpb"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/common/server/mongo"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found, using system environment variables")
	}
	config.MustInit()
}

func main() {
	cfg := config.Get()
	serviceName := cfg.ServerName
	if serviceName == "" {
		serviceName = "audit"
	}
	if err := logger.Init(&logger.Config{Level: cfg.LogLevel, ServiceName: serviceName}); err != nil {
		logger.Warn("logger init failed; the default logger will be used")
	}

	if err := cache.Initialize(cfg.ReadRedis, cfg.WriteRedis, cfg.RedisHealthCheckInterval); err != nil {
		logger.WithError(err).Error("redis init failed")
	}
	cacheClient, err := cache.GetClient()
	if err != nil {
		logger.WithError(err).Fatal("redis client not found")
	}
	if cfg.JWTConfig.AccessSecret == "" {
		logger.Fatal("JWT access secret is empty")
	}
	tokenIssuer := jwt.NewJWTIssuer(config.JWTConfig{
		AccessSecret:  cfg.JWTConfig.AccessSecret,
		RefreshSecret: cfg.JWTConfig.RefreshSecret,
		AccessTTL:     cfg.JWTConfig.AccessTTL,
		RefreshTTL:    cfg.JWTConfig.RefreshTTL,
	}, jwt.NewTokenCacheRepository(cacheClient.W()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		logger.WithError(err).Fatal("mongo connect")
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := mongoClient.Close(closeCtx); err != nil {
			logger.WithError(err).Error("mongo close")
		}
	}()

	application := service.NewApplication(ctx, service.ApplicationDependencies{MongoClient: mongoClient})

	discoveryConfig := cfg.ServiceDiscoveryConfig
	consulClient, err := consulx.NewClient(&consulx.Config{
		Address: fmt.Sprintf("%s:%d", discoveryConfig.Host, discoveryConfig.Port),
		Token:   discoveryConfig.ACTToken,
		Timeout: discoveryConfig.Timeout,
	})
	if err != nil {
		logger.WithError(err).Fatal("consul client")
	}
	instanceID := fmt.Sprintf("%s-%s-%d", serviceName, cfg.ServerHost, cfg.GRPCPort)
	registrar := registry.NewRegistrar(consulClient, registry.ServiceInfo{
		Name: serviceName,
		ID:   instanceID,
		Host: cfg.ServerHost,
		Port: cfg.GRPCPort,
	}, discoveryConfig.Timeout)

	httpErrCh := server.RunHTTPServerWithLifecycle(ctx, cfg.GetServerAddr(), func(router *gin.Engine) {
		middlewares := []protos.MiddlewareFunc{
			protos.MiddlewareFunc(server.AuthMiddleware(tokenIssuer)),
			protos.MiddlewareFunc(server.LoggerMiddleware()),
			protos.MiddlewareFunc(server.RateLimitMiddleware()),
		}
		protos.RegisterHandlersWithOptions(router, HTTPServer{app: application}, protos.GinServerOptions{
			BaseURL:     "/api",
			Middlewares: middlewares,
			ErrorHandler: func(c *gin.Context, err error, statusCode int) {
				logger.WithError(err).Error("audit http")
				server.AbortWithStatus(c, server.CodeInternalServerError, statusCode)
			},
		})
	}, server.WithHTTPInstanceID(instanceID))

	grpcErrCh := server.RunGRPCServerWithLifecycle(ctx, func(srv *grpc.Server) {
		auditpb.RegisterAuditServiceServer(srv, protos.NewGRPCServer(application))
	}, registrar, server.WithGRPCAuthParser(tokenIssuer), server.WithGRPCInstanceID(instanceID))

	logger.Info("audit service is ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-quit:
		logger.Info("received shutdown signal: " + sig.String())
	case err := <-grpcErrCh:
		logger.WithError(err).Error("grpc server stopped")
	case err := <-httpErrCh:
		logger.WithError(err).Error("http server stopped")
	}

	cancel()
	shutdownComplete := make(chan struct{})
	go func() {
		for range grpcErrCh {
		}
		for range httpErrCh {
		}
		close(shutdownComplete)
	}()
	select {
	case <-shutdownComplete:
		logger.Info("audit service stopped")
	case <-time.After(45 * time.Second):
		logger.Warn("shutdown timeout, forcing exit")
	}
	cache.CloseAll()
}
