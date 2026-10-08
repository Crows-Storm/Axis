package main

import (
	"github.com/Crows-Storm/Axis/common/bootstrap"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/genproto/ledgerpb"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/ledger/protos"
	"github.com/Crows-Storm/Axis/ledger/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg := bootstrap.Load("ledger")
	proc, err := bootstrap.Open(cfg)
	if err != nil {
		logger.WithError(err).Error("startup failed")
		return
	}
	defer proc.Close()

	application, cleanup := service.NewApplication(proc.Context(), proc.Cache)
	defer cleanup()

	proc.Serve(func(router *gin.Engine) {
		protos.RegisterHandlersWithOptions(router, HTTPServer{app: application}, protos.GinServerOptions{
			BaseURL: "/api",
			ErrorHandler: func(c *gin.Context, err error, statusCode int) {
				server.ErrorWithHttpCode(c, server.CodeInternalServerError, statusCode)
			},
		})
	}, func(srv *grpc.Server) {
		ledgerpb.RegisterLedgerServiceServer(srv, protos.NewGRPCServer(application))
	})
}
