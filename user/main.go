package main

import (
	"github.com/Crows-Storm/Axis/common/bootstrap"
	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/decorator"
	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/kafka"
	"github.com/Crows-Storm/Axis/common/server/store"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/Crows-Storm/Axis/user/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg := bootstrap.Load("user")
	proc, err := bootstrap.Open(cfg)
	if err != nil {
		logger.WithError(err).Error("startup failed")
		return
	}
	defer proc.Close()

	st, err := store.Open(cfg.DbConfig)
	if err != nil {
		logger.WithError(err).Error("database open failed")
		return
	}
	defer closeStore(st)

	issuer, err := proc.TokenIssuer()
	if err != nil {
		logger.WithError(err).Error("jwt issuer failed")
		return
	}

	authz := rbac.NewAuthorizer(rbac.NewRedis(proc.Cache), nil)
	if err := authz.PublishRoutes(proc.Context(), rbac.UserRoutes()); err != nil {
		logger.WithError(err).Error("publish rbac routes failed")
	}

	bus, err := openPublisher(cfg.Kafka, cfg.ServerName)
	if err != nil {
		logger.WithError(err).Error("kafka publisher failed")
		return
	}

	application, cleanup := service.NewApplication(proc.Context(), service.ApplicationDependencies{
		Store:       st,
		CacheClient: proc.Cache,
		Issuer:      issuer,
		EventBus:    bus,
	})
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
				logger.Error("API error", "path", c.FullPath(), "error", err, "status", statusCode)
				server.AbortWithStatus(c, server.CodeInternalServerError, statusCode)
			},
		})
	}, func(srv *grpc.Server) {
		userpb.RegisterUserServiceServer(srv, protos.NewGRPCServer(application))
	}, bootstrap.WithAuth(issuer))
}

func openPublisher(cfg config.KafkaConfig, serviceName string) (decorator.EventBus, error) {
	if !cfg.Enabled() {
		logger.Info("kafka publisher disabled")
		return nil, nil
	}
	return kafka.OpenPublisher(cfg, serviceName, true)
}

func closeStore(st *store.Store) {
	if err := st.Close(); err != nil {
		logger.WithError(err).Error("database close failed")
	}
}
