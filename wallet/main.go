package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/genproto/walletpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/common/server/store"
	"github.com/Crows-Storm/Axis/wallet/protos"
	"github.com/Crows-Storm/Axis/wallet/service"
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
	logger.Info("║           🔥 AXIS-WALLET - Universal Kanban System           ║")
	logger.Info("╚════════════════════════════════════════════════════════════╝")

	if err := redis.Initialize(cfg.ReadRedis, cfg.WriteRedis, cfg.RedisHealthCheckInterval); err != nil {
		logger.Error(err, "Failed to init redis")
	}

	logger.Info("✅ Redis Initialization complete")

	cacheClient, err := redis.GetClient()
	if err != nil {
		logger.WithError(err).Fatal("Cache redis instance not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, cleanup := service.NewApplication(ctx, cacheClient)
	defer cleanup()

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
	defer st.Close()

	go server.RunGRPCServer(func(server *grpc.Server) {
		walletServiceServer := protos.NewGRPCServer(application)
		walletpb.RegisterWalletServiceServer(server, walletServiceServer)
	})

	go server.RunHTTPServer(cfg.GetServerAddr(), func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HTTPServer{
			app: application, // inject application
		}, protos.GinServerOptions{
			BaseURL:      "/api/v1/wallet",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
		logger.Println("Start Successfully" + serviceName)
	})

	logger.Info("Service is ready 🚀")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.WithField("signal", sig.String()).Warn("Received shutdown signal")

	logger.Info("Shutting down Redis connections...")
	defer redis.CloseAll()

	defer cleanup()

	logger.Info("Graceful shutdown completed ✅")
}
