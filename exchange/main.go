package main

import (
	"github.com/Crows-Storm/Axis/common/bootstrap"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/genproto/exchangepb"
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/exchange/protos"
	"github.com/Crows-Storm/Axis/exchange/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg := bootstrap.Load("exchange")
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

	var appDependencies = service.ApplicationDependencies{
		CacheClient: proc.Cache,
		Issuer:      issuer,
	}
	application, cleanup := service.NewApplication(proc.Context(), appDependencies)
	defer cleanup()

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
				logger.WithError(err).Error("exchange http")
				server.AbortWithStatus(c, server.CodeInternalServerError, statusCode)
			},
		})
	}, func(srv *grpc.Server) {
		exchangepb.RegisterExchangeServiceServer(srv, protos.NewGRPCServer(application))
	}, bootstrap.WithAuth(issuer))
}
