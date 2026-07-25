package main

import (
	"context"
	"log"

	"github.com/Crows-Storm/Axis/auth/protos"
	"github.com/Crows-Storm/Axis/auth/service"
	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/genproto/authpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		panic("Init ViperConfig ERROR !!!")
	}
}

func main() {
	// init Redis
	if err := config.NewRedisConnect(); err != nil {
		log.Fatalf("[auth main] ❌ Redis init failed: %v", err)
	}
	defer config.CloseRedis()

	serviceName := viper.GetString("auth.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	application := service.NewApplication(ctx)

	defer cancel()

	// RunGRPCServer
	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		authServiceServer := protos.NewGRPCServer(application)
		authpb.RegisterAuthServiceServer(server, authServiceServer)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HttpServer{
			app: application,
		}, protos.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})

}
