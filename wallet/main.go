package main

import (
	"context"
	"log"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/genproto/walletpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/wallet/protos"
	"github.com/Crows-Storm/Axis/wallet/service"
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
	if err := config.NewRedisConnect(); err != nil {
		log.Fatalf("[wallet main] ❌ Redis init failed: %v", err)
	}
	defer config.CloseRedis()

	serviceName := viper.GetString("wallet.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	application := service.NewApplication(ctx)
	defer cancel()

	// Run GRPCServer
	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		walletServiceServer := protos.NewGRPCServer(application)
		walletpb.RegisterWalletServiceServer(server, walletServiceServer)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HTTPServer{
			app: application,
		}, protos.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
		log.Println("Start Successfully" + serviceName)
	})

}
