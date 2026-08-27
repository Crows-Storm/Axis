package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/gin-gonic/gin"
)

// RunHTTPServerWithLifecycle starts the HTTP server and supports graceful shutdown
func RunHTTPServerWithLifecycle(ctx context.Context, addr string, wrapper func(router *gin.Engine)) chan error {
	apiRouter := gin.New()

	apiRouter.Use(gin.Recovery())
	apiRouter.Use(RequestIDMiddleware())

	wrapper(apiRouter)

	apiRouter.Any("/api/ping", func(c *gin.Context) {
		Success(c, "pong")
	})

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
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
