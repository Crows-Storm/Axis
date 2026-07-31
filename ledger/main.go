package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/genproto/ledgerpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/ledger/protos"
	"github.com/Crows-Storm/Axis/ledger/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		panic("Init ViperConfig ERROR !!!")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("No .env.local.local.example file found, using system environment variables")
	}

	log := config.InitLogger("ledger")

	redisCfgs, err := config.LoadRedisConfigs()
	if err != nil {
		log.WithError(err).Fatal("Failed to load redis config")
	}
	if err := redis.Initialize(redisCfgs, log); err != nil {
		log.WithError(err).Fatal("Failed to init redis")
	}

	cacheClient, err := redis.Get("cache")
	if err != nil {
		log.WithError(err).Fatal("Cache redis instance not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, cleanup := service.NewApplication(ctx, log, cacheClient)
	defer cleanup()

	serviceName := viper.GetString("ledger.service-name")

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		ledgerServiceServer := protos.NewGRPCServer(application)
		ledgerpb.RegisterLedgerServiceServer(server, ledgerServiceServer)
	})

	go server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HTTPServer{
			app: application, // inject application
		}, protos.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
		log.Println("Start Successfully" + serviceName)
	})

	log.Info("Service is ready 🚀")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.WithField("signal", sig.String()).Warn("Received shutdown signal")

	// TODO: It seems ineffective
	log.Info("Shutting down Redis connections...")
	defer redis.CloseAll()

	defer cleanup()

	log.Info("Graceful shutdown completed ✅")
}
