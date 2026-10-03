package main

import (
	"github.com/Crows-Storm/Axis/wallet/app"
	"github.com/Crows-Storm/Axis/wallet/protos"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) GetWalletById(c *gin.Context, id uint64) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) GetWalletsByUserId(c *gin.Context, userId uint64, params protos.GetWalletsByUserIdParams) {
	//TODO implement me
	panic("implement me")
}
