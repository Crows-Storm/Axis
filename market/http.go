package main

import (
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/market/app"
	"github.com/Crows-Storm/Axis/market/protos"
	"github.com/gin-gonic/gin"
)

const serviceRole = "market-data"

type HTTPServer struct {
	app app.Application
}

var _ protos.ServerInterface = (*HTTPServer)(nil)

func (h HTTPServer) GetStatus(c *gin.Context) {
	// Market data: quotes, tickers, and candles. Not order routing and not chain watchers.
	server.Success(c, map[string]string{
		"service": "market",
		"role":    serviceRole,
	})
}
