package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/discovery/consulx"
	"github.com/Crows-Storm/Axis/common/discovery/registry"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/cache"
	"github.com/Crows-Storm/Axis/common/server/kafka"
	"github.com/Crows-Storm/Axis/common/server/kafka/outbox"
	"github.com/Crows-Storm/Axis/common/server/store"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/Crows-Storm/Axis/user/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("No .env file found, using system environment variables")
	}
	config.MustInit()
}

func main() {
	// get env global config
	cfg := config.Get()

	serviceName := cfg.ServerName
	err := logger.Init(&logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Warn("⚠️ Custom logger initialization failed; the default logger will be used instead ⚠️")
	}

	logger.Info("✅ Configuration loaded")
	logger.Info("✅ Logger Initialization complete")

	logger.Info("╔════════════════════════════════════════════════════════════╗")
	logger.Info("║           🔥 AXIS-USER - Universal Kanban System           ║")
	logger.Info("╚════════════════════════════════════════════════════════════╝")

	if err := cache.Initialize(cfg.ReadRedis, cfg.WriteRedis, cfg.RedisHealthCheckInterval); err != nil {
		logger.Error(err, "Failed to init redis")
	}

	logger.Info("✅ Redis Initialization complete")

	cacheClient, err := cache.GetClient()
	if err != nil {
		logger.WithError(err).Fatal("Cache redis instance not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init gorm and connection to db
	dbCfg := cfg.DbConfig
	logger.Infof("📋 Initializing database (%s)...", dbCfg.DBType)
	dbType := store.DBTypeSQLite
	if dbCfg.DBType == "mariadb" {
		dbType = store.DBTypeMaria
	}
	st, err := store.NewWithConfig(store.DBConfig{
		Type:     dbType,
		Path:     dbCfg.DBPath,
		Host:     dbCfg.DBHost,
		Port:     dbCfg.DBPort,
		User:     dbCfg.DBLoginId,
		Password: dbCfg.DBPassword,
		DBName:   dbCfg.DBSchema,
		SSLMode:  dbCfg.DBSslMode,
	})

	if err != nil {
		logger.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer func(st *store.Store) {
		err := st.Close()
		if err != nil {
			logger.Errorf("Failed to Close database connection: %v", err)
		}
	}(st)

	// TODO: need improve
	tokenRepository := jwt.NewTokenCacheRepository(cacheClient.W())
	if cfg == nil || cfg.JWTConfig.AccessSecret == "" {
		logger.Panicf("config.JWTConfig not init: cfg=%v", cfg)
		panic("config.JWTConfig not init!!!")
	}
	// init jwt issuer
	jwtConfig := cfg.JWTConfig
	tokenIssuer := jwt.NewJWTIssuer(config.JWTConfig{
		AccessSecret:  jwtConfig.AccessSecret,
		RefreshSecret: jwtConfig.RefreshSecret,
		AccessTTL:     jwtConfig.AccessTTL,
		RefreshTTL:    jwtConfig.RefreshTTL,
	}, tokenRepository)
	application, cleanup := service.NewApplication(ctx, service.ApplicationDependencies{
		Store:       st,
		CacheClient: cacheClient,
		Issuer:      tokenIssuer,
	})
	defer cleanup()

	discoveryConfig := cfg.ServiceDiscoveryConfig
	consulClient, err := consulx.NewClient(&consulx.Config{
		Address: fmt.Sprintf("%s:%d", discoveryConfig.Host, discoveryConfig.Port),
		Token:   discoveryConfig.ACTToken,
		Timeout: discoveryConfig.Timeout,
	})
	if err != nil {
		logger.Error("init consul failed", "error", err)
		os.Exit(1)
	}

	// register self to consul, This instanceId is A unique identifier for a service instance, used for service auditing/monitoring.
	instanceId := fmt.Sprintf("%s-%s-%d", serviceName, cfg.ServerHost, cfg.GRPCPort)
	registrar := registry.NewRegistrar(consulClient, registry.ServiceInfo{
		Name: serviceName,
		ID:   instanceId,
		Host: cfg.ServerHost,
		Port: cfg.GRPCPort,
	})

	// TODO: init and connect to kafka
	producerCfg := kafka.DefaultProducerConfig(cfg.KafkaBrokers)
	producerCfg.ClientID = "order-service-producer"
	producerCfg.TopicPrefix = "order"
	producer, err := kafka.NewProducer(producerCfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to init Kafka producer")
	}

	// ──── 2. 初始化序列化器 & Topic 路由 ────
	serializer := kafka.NewSerializer()
	serializer.Register("OrderCreated", func() event.DomainEvent { return &events.OrderCreated{} })
	serializer.Register("OrderCancelled", func() event.DomainEvent { return &events.OrderCancelled{} })

	// router prefix
	router := kafka.NewTopicRouter(serviceName)

	// ──── 3. 初始化 EventPublisher (EventBus 实现) ────
	publisher := kafka.NewEventPublisher(producer, serializer, router, logger.Log)
	publisher.Use(kafka.TracingMiddleware())
	publisher.Use(kafka.CorrelationIDMiddleware())
	publisher.Use(kafka.AuditMiddleware("audit"))

	// ──── 4. 初始化 Outbox ────
	outboxRepo := outbox.NewPostgresOutboxRepository(db)
	outboxProcessor := outbox.NewProcessor(outboxRepo, publisher, 2*time.Second, 100, logger)
	outboxProcessor.Start()

	// ──── 5. 初始化 Consumer ────
	dispatcher := event.NewEventDispatcher()

	// 注册消费端 handler
	dispatcher.Register(&handlers.PaymentCompletedHandler{...})
	dispatcher.Register(&handlers.InventoryReservedHandler{...})

	consumerCfg := kafka.DefaultConsumerConfig(cfg.KafkaBrokers, "order-service-group")
	consumerCfg.Topics = []string{
		"payment-service.payment-completed",
		"inventory-service.inventory-reserved",
	}
	consumerCfg.DeadLetterTopic = "order-service.dlq"

	consumer, err := kafka.NewConsumer(consumerCfg, dispatcher, producer, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to init Kafka consumer")
	}
	if err := consumer.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start Kafka consumer")
	}

	// ──── 6. 注入到 Application Service ────
	orderAppSvc := application.NewOrderService(publisher, outboxRepo)

	// ──── 7. 健康检查 ────
	healthChecker := kafka.NewHealthChecker(cfg.KafkaBrokers)
	// 注册到 HTTP health endpoint...

	// ──── 8. 优雅关闭 ────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 按依赖顺序关闭
	consumer.Stop()
	outboxProcessor.Stop()
	publisher.Close()

	_ = shutdownCtx
	logger.Info("Server exited gracefully")

	httpErrCh := server.RunHTTPServerWithLifecycle(ctx, cfg.GetServerAddr(), func(router *gin.Engine) {
		middlewares := []protos.MiddlewareFunc{
			protos.MiddlewareFunc(server.RequestIDMiddleware()),
			protos.MiddlewareFunc(server.LoggerMiddleware()),
			protos.MiddlewareFunc(server.AuthMiddleware(tokenIssuer)),
			protos.MiddlewareFunc(server.RateLimitMiddleware()),
			protos.MiddlewareFunc(server.PermissionMiddleware()),
		}

		protos.RegisterHandlersWithOptions(router, HTTPServer{
			app: application,
		}, protos.GinServerOptions{
			BaseURL:     "/api",
			Middlewares: middlewares,
			ErrorHandler: func(c *gin.Context, err error, statusCode int) {
				logger.Error("API error", "path", c.FullPath(), "error", err, "status", statusCode)
				server.AbortWithStatus(c, server.CodeInternalServerError, statusCode)
			},
		})
		logger.Info("HTTP routes registered successfully")
	})

	grpcErrCh := server.RunGRPCServerWithLifecycle(
		ctx,
		func(server *grpc.Server) {
			userServiceServer := protos.NewGRPCServer(application)
			userpb.RegisterUserServiceServer(server, userServiceServer)
		},
		registrar,
		// Restore the principal propagated by upstream services (e.g. auth)
		// so business handlers can read it via principal.FromContext(ctx).
		server.WithGRPCAuthParser(tokenIssuer),
	)

	logger.Info("Service is ready 🚀")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("received shutdown signal", "signal", sig.String())
	case err := <-grpcErrCh:
		logger.Error("gRPC server error, initiating shutdown", "error", err)
	case err := <-httpErrCh:
		logger.Error("HTTP server error, initiating shutdown", "error", err)
	}

	logger.Info("initiating graceful shutdown...")

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
		logger.Info("all servers stopped gracefully")
	case <-time.After(45 * time.Second):
		logger.Warn("shutdown timeout, forcing exit")
	}

	logger.Info("cleaning up application resources...")
	cleanup()

	logger.Info("closing Redis connections...")
	cache.CloseAll()

	logger.Info("graceful shutdown completed ✅")
}
