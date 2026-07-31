package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/redis"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/Crows-Storm/Axis/user/service"
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

	serviceName := viper.GetString("user.service-name")

	log := config.InitLogger(serviceName)

	log.Info("╔════════════════════════════════════════════════════════════╗")
	log.Info("║             💥 AXIS - MY PRODUCTIVITY TOOLS                ║")
	log.Info("╚════════════════════════════════════════════════════════════╝")

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

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		userServiceServer := protos.NewGRPCServer(application)
		userpb.RegisterUserServiceServer(server, userServiceServer)
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
