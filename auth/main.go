package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Crows-Storm/Axis/auth/protos"
	"github.com/Crows-Storm/Axis/auth/service"
	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/genproto/authpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/redis"
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
		log.Warn("No .env.local.local file found, using system environment variables")
	}

	log := config.InitLogger("auth")

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

	serviceName := viper.GetString("auth.service-name")

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		authServiceServer := protos.NewGRPCServer(application)
		authpb.RegisterAuthServiceServer(server, authServiceServer)
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

	// TODO 2026/07/28 20:08: Need refactor GRPC & HTTP Server, This allows for greater flexibility and customization, even when using go gen.
	// Received quit signal to make grpc server on HealthCheckResponse_NOT_SERVING status
	//healthServer := health.NewServer()
	//healthpb.RegisterHealthServer(grpcServer, healthServer)
	//healthServer.SetServingStatus(serviceName, healthpb.HealthCheckResponse_NOT_SERVING)
	//
	//log.Info("Shutting down gRPC server...")
	//grpcServer.GracefulStop()
	//
	//// Graceful Stop HTTP (Waiting for the request being processed to complete, at most, waiting: 10s)
	//log.Info("Shutting down HTTP server...")
	//shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer shutdownCancel()
	//if err := httpServer.Shutdown(shutdownCtx); err != nil {
	//	log.WithError(err).Error("HTTP server forced to shutdown")
	//}

	// TODO: It seems ineffective
	log.Info("Shutting down Redis connections...")
	defer redis.CloseAll()

	defer cleanup()

	log.Info("Graceful shutdown completed ✅")
}
