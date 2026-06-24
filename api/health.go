// Package api exposes the application's HTTP surface, including the health
// diagnostics endpoint.
package api

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-production-grade/health"
)

// startedAt records when the process came up, reported by the health endpoint
// as an uptime reference.
var startedAt = time.Now().Format(time.RFC3339)

// API holds the dependencies needed to serve the HTTP surface.
type API struct {
	health  *health.Health
	version string
}

// New builds an API backed by the given health registry.
func New(checks *health.Health, opts ...Option) *API {
	opt := option{version: "v0.1.0"}
	for _, o := range opts {
		o(&opt)
	}
	return &API{health: checks, version: opt.version}
}

// Health holds the application's overall status and the status of each
// dependency it relies on.
type Health struct {
	// Version is the git tag of the running build.
	Version string `json:"version"`
	// Host is the host the application is running on.
	Host string `json:"host"`
	// Commit is the git commit the build was cut from.
	Commit string `json:"commit"`
	// TIER is the environment the application is deployed to.
	TIER string `json:"tier"`
	// Status is the overall status of the system: "ok" or "fail".
	Status string `json:"status"`
	// Started is the time the application started, RFC3339.
	Started string `json:"startedAt"`
	// Dependencies is the per-dependency status keyed by name.
	Dependencies map[string]any `json:"dependencies"`
}

// Healthy reports whether the overall status is "ok".
func (h Health) Healthy() bool { return h.Status == string(health.StatusOK) }

// Health runs every registered dependency check and assembles the report.
func (a *API) Health(ctx context.Context) Health {
	report := a.health.Check(ctx)

	host, _ := os.Hostname()
	deps := make(map[string]any, len(report.Checks))
	for name, check := range report.Checks {
		deps[name] = check
	}

	return Health{
		Version:      a.version,
		Host:         host,
		Commit:       os.Getenv("COMMIT_ID"),
		TIER:         os.Getenv("TIER"),
		Status:       string(report.Status),
		Started:      startedAt,
		Dependencies: deps,
	}
}

// HealthHandler is a gin handler that serves the health report. It returns 200
// when healthy and 503 when any dependency check fails, so it doubles as a
// readiness probe for load balancers and orchestrators.
func (a *API) HealthHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := a.Health(ctx.Request.Context())
		code := http.StatusOK
		if !h.Healthy() {
			code = http.StatusServiceUnavailable
		}
		ctx.JSON(code, h)
	}
}
