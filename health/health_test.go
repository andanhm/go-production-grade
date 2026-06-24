package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-production-grade/health"
	"github.com/stretchr/testify/assert"
)

func TestHealth_AllOK(t *testing.T) {
	h := health.New()
	h.Register("db", func(context.Context) error { return nil })
	h.Register("cache", func(context.Context) error { return nil })

	report := h.Check(context.Background())

	assert.Equal(t, health.StatusOK, report.Status)
	assert.Len(t, report.Checks, 2)
	assert.Equal(t, health.StatusOK, report.Checks["db"].Status)
	assert.NotEmpty(t, report.Checks["db"].Latency)
}

func TestHealth_OneFailFailsOverall(t *testing.T) {
	h := health.New()
	h.Register("db", func(context.Context) error { return nil })
	h.Register("cache", func(context.Context) error { return errors.New("connection refused") })

	report := h.Check(context.Background())

	assert.Equal(t, health.StatusFail, report.Status)
	assert.Equal(t, health.StatusOK, report.Checks["db"].Status)
	assert.Equal(t, health.StatusFail, report.Checks["cache"].Status)
	assert.Equal(t, "connection refused", report.Checks["cache"].Error)
}

func TestHealth_Empty(t *testing.T) {
	report := health.New().Check(context.Background())

	assert.Equal(t, health.StatusOK, report.Status)
	assert.Empty(t, report.Checks)
}

// TestHealth_SlowCheckTimesOut verifies a probe that ignores its context is
// bounded by the per-check timeout and reported as failed, rather than hanging.
func TestHealth_SlowCheckTimesOut(t *testing.T) {
	h := health.New(health.WithTimeout(50 * time.Millisecond))
	h.Register("slow", func(ctx context.Context) error {
		<-ctx.Done() // honour the bounded context
		return ctx.Err()
	})

	start := time.Now()
	report := h.Check(context.Background())
	elapsed := time.Since(start)

	assert.Equal(t, health.StatusFail, report.Status)
	assert.Less(t, elapsed, time.Second, "Check must return promptly, not hang")
}

// TestHealth_RequestContextCancelled verifies the report returns promptly when
// the caller's context is cancelled before checks settle.
func TestHealth_RequestContextCancelled(t *testing.T) {
	h := health.New(health.WithTimeout(time.Hour))
	h.Register("wedged", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	report := h.Check(ctx)
	elapsed := time.Since(start)

	assert.Equal(t, health.StatusFail, report.Status)
	assert.Equal(t, health.StatusFail, report.Checks["wedged"].Status)
	assert.Less(t, elapsed, time.Second)
}
