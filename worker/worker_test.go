package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-production-grade/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPool_DrainsAllJobs verifies that with a generous deadline, every
// submitted job runs to completion before Shutdown returns.
func TestPool_DrainsAllJobs(t *testing.T) {
	const jobs = 200

	var processed atomic.Int64
	p := worker.New(4, func(_ context.Context, n int) {
		processed.Add(int64(n))
	})
	p.Start(context.Background())

	var want int64
	for i := 1; i <= jobs; i++ {
		want += int64(i)
		require.NoError(t, p.Submit(i))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	assert.NoError(t, p.Shutdown(ctx))
	assert.Equal(t, want, processed.Load(), "every queued job should have drained")
}

// TestPool_ShutdownDeadline verifies that when handlers are slower than the
// drain deadline, Shutdown returns the deadline error promptly instead of
// hanging.
func TestPool_ShutdownDeadline(t *testing.T) {
	release := make(chan struct{})
	p := worker.New(2, func(_ context.Context, _ int) {
		<-release // block until the test lets go
	}, worker.WithBuffer(16))
	p.Start(context.Background())

	// Saturate both workers so they are stuck inside the slow handler.
	require.NoError(t, p.Submit(1))
	require.NoError(t, p.Submit(2))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := p.Shutdown(ctx)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed, time.Second, "Shutdown must return promptly on deadline, not hang")

	close(release) // unblock the stragglers so they exit cleanly
}

// TestPool_SubmitAfterShutdown verifies Submit reports ErrClosed once the pool
// is shutting down.
func TestPool_SubmitAfterShutdown(t *testing.T) {
	p := worker.New(2, func(_ context.Context, _ int) {})
	p.Start(context.Background())

	require.NoError(t, p.Shutdown(context.Background()))

	assert.ErrorIs(t, p.Submit(1), worker.ErrClosed)
}

// TestPool_ShutdownIdempotent verifies Shutdown can be called repeatedly
// without panicking on a double close of the job channel.
func TestPool_ShutdownIdempotent(t *testing.T) {
	p := worker.New(2, func(_ context.Context, _ int) {})
	p.Start(context.Background())

	assert.NoError(t, p.Shutdown(context.Background()))
	assert.NoError(t, p.Shutdown(context.Background()))
}

// TestPool_ForcedAbort verifies that cancelling the Start context stops the
// workers even while a slow handler is running.
func TestPool_ForcedAbort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	p := worker.New(2, func(c context.Context, _ int) {
		<-c.Done() // honour the work context
	})
	p.Start(ctx)

	require.NoError(t, p.Submit(1))
	require.NoError(t, p.Submit(2))

	cancel() // forced abort

	// Shutdown should now drain quickly because the workers observe ctx.Done().
	drainCtx, drainCancel := context.WithTimeout(context.Background(), time.Second)
	defer drainCancel()
	assert.NoError(t, p.Shutdown(drainCtx))
}
