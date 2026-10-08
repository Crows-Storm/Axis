package main

import (
	"context"
	"time"

	"github.com/Crows-Storm/Axis/audit/protos"
	"github.com/Crows-Storm/Axis/audit/service"
	"github.com/Crows-Storm/Axis/common/bootstrap"
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/domain/event"
	"github.com/Crows-Storm/Axis/common/genproto/auditpb"
	"github.com/Crows-Storm/Axis/common/rbac"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/common/server/kafka"
	"github.com/Crows-Storm/Axis/common/server/mongo"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cfg := bootstrap.Load("audit")
	proc, err := bootstrap.Open(cfg)
	if err != nil {
		logger.WithError(err).Error("startup failed")
		return
	}
	defer proc.Close()

	mongoClient, err := mongo.Connect(proc.Context(), cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		logger.WithError(err).Error("mongo connect failed")
		return
	}
	defer closeMongo(mongoClient)

	issuer, err := proc.TokenIssuer()
	if err != nil {
		logger.WithError(err).Error("jwt issuer failed")
		return
	}

	authz := rbac.NewAuthorizer(rbac.NewRedis(proc.Cache), nil)
	if err := authz.PublishRoutes(proc.Context(), rbac.AuditRoutes()); err != nil {
		logger.WithError(err).Error("publish rbac routes failed")
	}

	application := service.NewApplication(proc.Context(), service.ApplicationDependencies{MongoClient: mongoClient})

	if cfg.Kafka.Enabled() {
		group, err := kafka.OpenConsumer(proc.Context(), cfg.Kafka, cfg.ServerName, []string{event.AuditTopic}, func(ctx context.Context, delivery kafka.Delivery) error {
			return service.ConsumeRecord(ctx, application, delivery)
		})
		if err != nil {
			logger.WithError(err).Error("kafka consumer failed")
			return
		}
		defer group.Close()
	} else {
		logger.Info("kafka consumer disabled")
	}

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
				logger.WithError(err).Error("audit http")
				server.AbortWithStatus(c, server.CodeInternalServerError, statusCode)
			},
		})
	}, func(srv *grpc.Server) {
		auditpb.RegisterAuditServiceServer(srv, protos.NewGRPCServer(application))
	}, bootstrap.WithAuth(issuer))
}

func closeMongo(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Close(ctx); err != nil {
		logger.WithError(err).Error("mongo close failed")
	}
}
