package server

import (
	"strings"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/gin-gonic/gin"
)

var publicPaths = map[string]bool{
	"/api/login":    true,
	"/api/register": true,
	"/api/ping":     true,
}

func RequestContextMiddleware(instanceID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		builder := contextx.ExtractFromGinHeaders(c)
		builder.WithInstanceID(instanceID)
		ctx := builder.Build(c.Request.Context())

		if contextx.FromContext(ctx).RequestID() == "" {
			ErrorWithCode(c, CodeBadRequest)
			return
		}

		c.Request = c.Request.WithContext(ctx)
		contextx.InjectIntoGinHeaders(c, ctx)
		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		r := contextx.FromGinContext(c)

		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"request_id", r.RequestID(),
			"correlation_id", r.CorrelationID(),
			"span_id", r.SpanID(),
			"instance_id", r.InstanceID(),
			"user_id", r.UserID(),
			"client_ip", r.ClientIP(),
		)
	}
}

func AuthMiddleware(jwtIssuer *jwt.JWTIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicPaths[c.FullPath()] {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if token == "" {
			ErrorWithCode(c, CodeInvalidToken)
			return
		}
		token, err := jwtIssuer.Preprocessing(c.Request.Context(), token)
		if err != nil {
			ErrorWithCode(c, CodeInvalidToken)
			return
		}
		parse, err := jwtIssuer.Parse(c.Request.Context(), token)
		if err != nil {
			ErrorWithCode(c, CodeInvalidToken)
			return
		}
		ctx := principal.WithPrincipal(c.Request.Context(), parse)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicPaths[c.FullPath()] || c.FullPath() == "/api/logout" || c.FullPath() == "/api/" {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		pathPermissionCode := toPermissionCode(path)

		val, exists := c.Get("permission")
		if !exists {
			ErrorWithCode(c, CodeInsufficientPermissions)
			return
		}

		permission, ok := val.(map[string]struct{})
		if !ok {
			ErrorWithCode(c, CodeInsufficientPermissions)
			return
		}

		if _, has := permission[pathPermissionCode]; has {
			c.Next()
			return
		}

		ErrorWithCode(c, CodeInsufficientPermissions)
	}
}

func toPermissionCode(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	return strings.ReplaceAll(trimmed, "/", ":")
}
