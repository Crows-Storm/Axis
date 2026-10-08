package main

import (
	"github.com/Crows-Storm/Axis/chain/app"
	"github.com/Crows-Storm/Axis/chain/protos"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
)

const serviceRole = "blockchain-integration"

type HTTPServer struct {
	app app.Application
}

var _ protos.ServerInterface = (*HTTPServer)(nil)

func (h HTTPServer) GetStatus(c *gin.Context) {
	// Blockchain integration and listeners. RPC, subscriptions, confirmations. No keys in source; trading stays in exchange.
	server.Success(c, map[string]string{
		"service": "chain",
		"role":    serviceRole,
	})
}
