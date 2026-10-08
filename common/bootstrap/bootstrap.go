package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/discovery/consulx"
	"github.com/Crows-Storm/Axis/common/discovery/registry"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const shutdownTimeout = 45 * time.Second

// Process is the shared process lifetime: config, Redis, Consul, and the root context.
type Process struct {
	Config     *config.Config
	Cache      cache.RueidisClient
	Registrar  *registry.Registrar
	InstanceID string

	ctx    context.Context
	cancel context.CancelFunc
}

type serveConfig struct {
	tokenParser server.TokenParser
}

// Option configures HTTP/gRPC servers started by Serve.
type Option func(*serveConfig)

// WithAuth parses bearer tokens on the gRPC server.
func WithAuth(parser server.TokenParser) Option {
	return func(cfg *serveConfig) {
		cfg.tokenParser = parser
	}
}

// Load reads a local .env when one exists, then initializes config and the logger.
// fallbackName is used when SERVER_NAME is empty.
func Load(fallbackName string) *config.Config {
	if err := godotenv.Load(); err != nil {
		logger.Warn("no .env file found, using process environment")
	}
	config.MustInit()
	cfg := config.Get()
	if cfg.ServerName == "" {
		cfg.ServerName = fallbackName
	}
	if err := logger.Init(&logger.Config{Level: cfg.LogLevel, ServiceName: cfg.ServerName}); err != nil {
		logger.Warn("logger init failed; the default logger will be used")
	}
	logger.Info("service starting",
		"service", cfg.ServerName,
		"http", cfg.GetServerAddr(),
		"grpc", cfg.GetGRPCAddr(),
	)
	return cfg
}

// Open connects Redis and Consul. On failure any connection already opened is closed.
func Open(cfg *config.Config) (*Process, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if err := cache.Initialize(cfg.ReadRedis, cfg.WriteRedis, cfg.RedisHealthCheckInterval); err != nil {
		return nil, fmt.Errorf("redis init: %w", err)
	}
	cacheClient, err := cache.GetClient()
	if err != nil {
		cache.CloseAll()
		return nil, fmt.Errorf("redis client: %w", err)
	}

	discovery := cfg.ServiceDiscoveryConfig
	consulClient, err := consulx.NewClient(&consulx.Config{
		Address: fmt.Sprintf("%s:%d", discovery.Host, discovery.Port),
		Token:   discovery.ACTToken,
		Timeout: discovery.Timeout,
	})
	if err != nil {
		cache.CloseAll()
		return nil, fmt.Errorf("consul client: %w", err)
	}

	instanceID := fmt.Sprintf("%s-%s-%d", cfg.ServerName, cfg.ServerHost, cfg.GRPCPort)
	ctx, cancel := context.WithCancel(context.Background())
	return &Process{
		Config: cfg,
		Cache:  cacheClient,
		Registrar: registry.NewRegistrar(consulClient, registry.ServiceInfo{
			Name: cfg.ServerName,
			ID:   instanceID,
			Host: cfg.ServerHost,
			Port: cfg.GRPCPort,
		}, discovery.Timeout),
		InstanceID: instanceID,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Context is cancelled when Serve shuts the process down.
func (p *Process) Context() context.Context {
	if p == nil || p.ctx == nil {
		return context.Background()
	}
	return p.ctx
}

// TokenIssuer builds a JWT issuer from config. The access secret must be set.
func (p *Process) TokenIssuer() (*jwt.JWTIssuer, error) {
	if p == nil || p.Config == nil || p.Config.JWTConfig.AccessSecret == "" {
		return nil, errors.New("JWT access secret is empty")
	}
	if p.Cache == nil {
		return nil, errors.New("redis client is nil")
	}
	return jwt.NewJWTIssuer(p.Config.JWTConfig, jwt.NewTokenCacheRepository(p.Cache.W())), nil
}

// Serve runs HTTP and gRPC until a signal or a server error, then waits for both to stop.
func (p *Process) Serve(registerHTTP func(*gin.Engine), registerGRPC func(*grpc.Server), opts ...Option) {
	cfg := serveConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	httpErrCh := server.RunHTTPServerWithLifecycle(
		p.ctx,
		p.Config.GetServerAddr(),
		registerHTTP,
		server.WithHTTPInstanceID(p.InstanceID),
	)
	grpcOpts := []server.GRPCServerOption{server.WithGRPCInstanceID(p.InstanceID)}
	if cfg.tokenParser != nil {
		grpcOpts = append(grpcOpts, server.WithGRPCAuthParser(cfg.tokenParser))
	}
	grpcErrCh := server.RunGRPCServerWithLifecycle(p.ctx, registerGRPC, p.Registrar, grpcOpts...)

	logger.Info("service is ready", "service", p.Config.ServerName)
	wait(p.cancel, httpErrCh, grpcErrCh, shutdownTimeout)
}

// Close cancels the process context and closes Redis.
func (p *Process) Close() {
	if p == nil {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	cache.CloseAll()
}

func wait(cancel context.CancelFunc, httpErrCh, grpcErrCh <-chan error, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		logger.Info("received shutdown signal", "signal", sig.String())
	case err := <-grpcErrCh:
		logger.WithError(err).Error("grpc server stopped")
	case err := <-httpErrCh:
		logger.WithError(err).Error("http server stopped")
	}

	logger.Info("shutting down")
	cancel()

	done := make(chan struct{})
	go func() {
		for range grpcErrCh {
		}
		for range httpErrCh {
		}
		close(done)
	}()
	select {
	case <-done:
		logger.Info("servers stopped")
	case <-time.After(timeout):
		logger.Warn("shutdown timeout, forcing exit")
	}
}
