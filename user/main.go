package main

import (
	"log"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	// RunGRPCServer

	// Run HttpServer
	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HTTPServer{
			app: application, // inject application
		})
	})

}
