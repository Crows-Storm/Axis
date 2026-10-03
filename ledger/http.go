package main

import (
	"github.com/Crows-Storm/Axis/ledger/app"
	"github.com/Crows-Storm/Axis/ledger/protos"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) GetLedgerByWalletId(c *gin.Context, id uint64, params protos.GetLedgerByWalletIdParams) {
	//TODO implement me
	panic("implement me")
}
