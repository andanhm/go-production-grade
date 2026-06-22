package signals_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/go-production-grade/signals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContext_SignalCancels verifies that delivering one of the watched signals
// to the current process cancels the derived context.
func TestContext_SignalCancels(t *testing.T) {
	ctx, stop := signals.Context(context.Background(), syscall.SIGTERM)
	defer stop()

	// The handler is registered by the call above, so raising SIGTERM here is
	// caught instead of terminating the test binary.
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))

	select {
	case <-ctx.Done():
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled after SIGTERM")
	}
}

// TestContext_ParentCancels verifies that cancelling the parent context (no
// signal involved) also cancels the derived context.
func TestContext_ParentCancels(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	ctx, stop := signals.Context(parent)
	defer stop()

	cancel()

	select {
	case <-ctx.Done():
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled after parent cancellation")
	}
}

// TestContext_DefaultSignals verifies that omitting signals falls back to the
// Default set (SIGINT, SIGTERM).
func TestContext_DefaultSignals(t *testing.T) {
	assert.Equal(t, []os.Signal{syscall.SIGINT, syscall.SIGTERM}, signals.Default)

	ctx, stop := signals.Context(context.Background())
	defer stop()

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled after SIGINT")
	}
}

// TestContext_StopReleases verifies that stop unregisters the handler and
// cancels the context (it is a context.CancelFunc) even when no signal is ever
// received, and that calling it twice is safe.
func TestContext_StopReleases(t *testing.T) {
	ctx, stop := signals.Context(context.Background())

	assert.NoError(t, ctx.Err(), "context should be live before stop")

	stop()
	stop() // idempotent

	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}
