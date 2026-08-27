package server

import (
	"strings"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/discovery/grpcx"
	"github.com/Crows-Storm/Axis/common/domain/principal"
	"github.com/Crows-Storm/Axis/common/jwt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

var publicPaths = map[string]bool{
	"/api/login":    true,
	"/api/register": true,
	"/api/ping":     true,
}

const bearerPrefix = "Bearer "

// RequestIDMiddleware just check has X-Request-ID and setting to header for gin.Context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			// cannot no request id
			ErrorWithCode(c, CodeBadRequest)
			return
		}
		c.Header("trace_id", traceID)
		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logger.Info("➡️[API request] ",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"trace_id", c.GetString("trace_id"),
		)
	}
}

// AuthMiddleware check token and setting to gin.Context: principalKey contextKey = "auth:principal"
func AuthMiddleware(jwt *jwt.JWTIssuer) gin.HandlerFunc {
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
		token, err := jwt.Preprocessing(c, token)
		if err != nil {
			ErrorWithCode(c, CodeInvalidToken)
			return
		}
		parse, err := jwt.Parse(c, token)
		if err != nil {
			ErrorWithCode(c, CodeInvalidToken)
			return
		}
		// setting principal to context
		ctx := principal.WithPrincipal(c.Request.Context(), parse)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// TODO
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//if !allowRequest(c.ClientIP()) {
		//	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		//		"code": 429, "message": "rate limit exceeded",
		//	})
		//	return
		//}
		c.Next()
	}
}

// TODO
func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicPaths[c.FullPath()] || "/api/logout" == c.FullPath() || "/api/" == c.FullPath() {
			c.Next()
			return
		}
		// e.g.: GET /user/api/				self information
		// e.g.: GET /user/api/123			by user id
		// e.g.: PUT /user/api/disable				is disable a user
		// e.g.: PUT /user/api/				is Update
		// e.g.: POST /user/api/				is Create
		// e.g.: DELETE /user/api/				is Soft delete
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
		return
	}
}

// RequestMetadataMiddleware bridges HTTP request headers into gRPC outgoing metadata,
// so the outbound interceptor (grpcx.metadataPropagationInterceptor) can propagate
// them to downstream gRPC services.
func RequestMetadataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			ErrorWithCode(c, CodeBadRequest)
			return
		}

		md := metadata.MD{}
		md.Set(grpcx.KeyRequestID, requestID)
		md.Set(grpcx.KeyTraceID, requestID)
		if auth := c.GetHeader("Authorization"); auth != "" {
			md.Set(grpcx.KeyAuthorization, auth)
		}

		ctx := grpcx.WithOutgoingMetadata(c.Request.Context(), md)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func toPermissionCode(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	result := strings.ReplaceAll(trimmed, "/", ":")
	return result
}
