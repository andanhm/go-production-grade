package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-production-grade/api"
	"github.com/go-production-grade/health"
	"github.com/go-production-grade/logging"
	"github.com/go-production-grade/signals"
)

func main() {
	logging.Init("go-production-grade")

	// Register a check per dependency. A "self" liveness check keeps the
	// server runnable without a database; wire real db/cache checks here as
	// connection handles become available (see api/README.md).
	checks := health.New(health.WithTimeout(2 * time.Second))
	checks.Register("self", func(context.Context) error { return nil })

	app := api.New(checks, api.WithVersion("v0.1.0"))

	router := gin.New()
	router.Use(gin.Recovery())

	v2 := router.Group("/v2")
	v2.GET("/health", app.HealthHandler())

	// Context cancelled on SIGINT/SIGTERM (Day 4).
	ctx, stop := signals.Context(context.Background())
	defer stop()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Serve in the background; ListenAndServe blocks until Shutdown is called.
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Critical("ERR.SERVER.START", err.Error(), logging.WithTraceError(err))
		}
	}()
	logging.Info("APP.STARTED", "http server listening on :8080")

	// Block until a termination signal arrives, then drain in-flight requests
	// within a bounded window (Day 5).
	<-ctx.Done()
	logging.Info("APP.SHUTDOWN", "draining in-flight requests")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Error("ERR.SERVER.SHUTDOWN", err.Error(), logging.WithTraceError(err))
		return
	}
	logging.Info("APP.STOPPED", "shutdown complete")
}
