package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/gin-gonic/gin"
)

// RunHTTPServerWithLifecycle starts the HTTP server and supports graceful shutdown
// ctx: The context used to receive shutdown signals
// addr: The listening address, e.g., ":8080"
// wrapper: Callback functions used to register routes
// Returns an error channel, allowing the caller to listen for startup errors
func RunHTTPServerWithLifecycle(ctx context.Context, addr string, wrapper func(router *gin.Engine)) chan error {
	apiRouter := gin.New()

	apiRouter.Any("/api/v1/ping", func(c *gin.Context) {
		Success(c, "pong")
	})

	wrapper(apiRouter)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      apiRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Info("HTTP server started", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			errCh <- err
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Info("HTTP server shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server forced shutdown", "error", err)
			errCh <- err
		} else {
			logger.Info("HTTP server stopped gracefully")
		}
		close(errCh)
	}()

	return errCh
}
