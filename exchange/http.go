package main

import (
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/exchange/app"
	"github.com/Crows-Storm/Axis/exchange/protos"
	"github.com/gin-gonic/gin"
)

const serviceRole = "exchange-integration"

type HTTPServer struct {
	app app.Application
}

var _ protos.ServerInterface = (*HTTPServer)(nil)

func (h HTTPServer) GetStatus(c *gin.Context) {
	// HTTP is health only. Other services call venue APIs over internal gRPC.
	server.Success(c, map[string]string{
		"service": "exchange",
		"role":    serviceRole,
	})
}
