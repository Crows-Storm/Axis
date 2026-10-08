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
	// Exchange integration: venue accounts, orders, and fills. Prices come from market; chain settlement from chain.
	server.Success(c, map[string]string{
		"service": "exchange",
		"role":    serviceRole,
	})
}
