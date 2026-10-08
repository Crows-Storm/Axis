package main

import (
	"github.com/Crows-Storm/Axis/chain/protos"
	"github.com/Crows-Storm/Axis/chain/service"
	"github.com/Crows-Storm/Axis/common/bootstrap"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/genproto/chainpb"
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg := bootstrap.Load("chain")
	proc, err := bootstrap.Open(cfg)
	if err != nil {
		logger.WithError(err).Error("startup failed")
		return
	}
	defer proc.Close()

	issuer, err := proc.TokenIssuer()
	if err != nil {
		logger.WithError(err).Error("jwt issuer failed")
		return
	}

	authz := rbac.NewAuthorizer(rbac.NewRedis(proc.Cache), nil)
	if err := authz.PublishRoutes(proc.Context(), rbac.StatusRoutes()); err != nil {
		logger.WithError(err).Error("publish rbac routes failed")
	}

	application := service.NewApplication(proc.Context())

	proc.Serve(func(router *gin.Engine) {
		middlewares := []protos.MiddlewareFunc{
			protos.MiddlewareFunc(server.AuthMiddleware(issuer)),
			protos.MiddlewareFunc(server.LoggerMiddleware()),
			protos.MiddlewareFunc(server.RateLimitMiddleware()),
			protos.MiddlewareFunc(server.PermissionMiddleware(authz)),
		}
		protos.RegisterHandlersWithOptions(router, HTTPServer{app: application}, protos.GinServerOptions{
			BaseURL:     "/api",
			Middlewares: middlewares,
			ErrorHandler: func(c *gin.Context, err error, statusCode int) {
				logger.WithError(err).Error("chain http")
				server.AbortWithStatus(c, server.CodeInternalServerError, statusCode)
			},
		})
	}, func(srv *grpc.Server) {
		chainpb.RegisterChainServiceServer(srv, protos.NewGRPCServer(application))
	}, bootstrap.WithAuth(issuer))
}
