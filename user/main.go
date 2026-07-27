package main

import (
	"context"
	"log"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/Crows-Storm/Axis/user/service"
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
		log.Fatalf("[user main] ❌ Redis init failed: %v", err)
	}

	defer config.CloseRedis()

	// Run Gin HttpServer and GRPCServer
	serviceName := viper.GetString("user.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	application := service.NewApplication(ctx)

	defer cancel()

	// RunGRPCServer
	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		userServiceServer := protos.NewGRPCServer(application)
		userpb.RegisterUserServiceServer(server, userServiceServer)
	})

	// Run HttpServer
	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HTTPServer{
			app: application, // inject application
		}, protos.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
		log.Println("Start Successfully" + serviceName)
	})
}
