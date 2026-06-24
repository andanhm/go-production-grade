package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-production-grade/api"
	"github.com/go-production-grade/health"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		checks     map[string]health.CheckFunc
		wantCode   int
		wantStatus string
	}{
		{
			name: "all dependencies healthy returns 200",
			checks: map[string]health.CheckFunc{
				"db":    func(context.Context) error { return nil },
				"cache": func(context.Context) error { return nil },
			},
			wantCode:   http.StatusOK,
			wantStatus: "ok",
		},
		{
			name: "one failing dependency returns 503",
			checks: map[string]health.CheckFunc{
				"db":    func(context.Context) error { return nil },
				"cache": func(context.Context) error { return errors.New("connection refused") },
			},
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: "fail",
		},
		{
			name:       "no dependencies returns 200",
			checks:     map[string]health.CheckFunc{},
			wantCode:   http.StatusOK,
			wantStatus: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checks := health.New()
			for name, fn := range tt.checks {
				checks.Register(name, fn)
			}
			a := api.New(checks, api.WithVersion("v1.2.3"))

			router := gin.New()
			router.GET("/v2/health", a.HealthHandler())

			req := httptest.NewRequest(http.MethodGet, "/v2/health", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)

			var body api.Health
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.wantStatus, body.Status)
			assert.Equal(t, "v1.2.3", body.Version)
			assert.Len(t, body.Dependencies, len(tt.checks))
		})
	}
}
