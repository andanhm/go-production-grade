package worker

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed is returned by Submit once the pool has begun shutting down and is
// no longer accepting new jobs.
var ErrClosed = errors.New("worker: pool is shutting down")

// Pool is a bounded set of background goroutines that process jobs of type T.
//
// Shutdown follows the same semantics as http.Server.Shutdown: it stops
// accepting new work and lets in-flight and already-queued jobs drain. The
// context passed to Start governs the *work* — cancelling it aborts running
// handlers — while the context passed to Shutdown governs only the *drain
// deadline*. The two are intentionally separate so a graceful drain does not
// abort the very jobs it is waiting on.
type Pool[T any] struct {
	handler func(context.Context, T)
	jobs    chan T
	size    int
	wg      sync.WaitGroup
	start   sync.Once

	mu     sync.RWMutex // guards closed and the close of jobs
	closed bool
}

// New creates a pool of size goroutines that invoke handler for each submitted
// job. A size below 1 is clamped to 1. By default the job queue is buffered to
// size; use WithBuffer to override.
func New[T any](size int, handler func(context.Context, T), opts ...Option) *Pool[T] {
	if size < 1 {
		size = 1
	}
	opt := option{buffer: size}
	for _, o := range opts {
		o(&opt)
	}
	if opt.buffer < 0 {
		opt.buffer = 0
	}
	return &Pool[T]{
		handler: handler,
		jobs:    make(chan T, opt.buffer),
		size:    size,
	}
}

// Start launches the worker goroutines. It is safe to call multiple times; only
// the first call has any effect. The provided context is propagated to every
// handler invocation; cancelling it forces running handlers to abort (assuming
// they honour the context) and the workers to exit.
func (p *Pool[T]) Start(ctx context.Context) {
	p.start.Do(func() {
		p.wg.Add(p.size)
		for range p.size {
			go p.worker(ctx)
		}
	})
}

func (p *Pool[T]) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			// Forced abort: the Start context was cancelled.
			return
		case job, ok := <-p.jobs:
			if !ok {
				// Graceful exit: the queue was closed by Shutdown and is
				// now fully drained.
				return
			}
			p.handler(ctx, job)
		}
	}
}

// Submit enqueues a job. It returns ErrClosed if the pool is shutting down.
//
// Submit applies backpressure: if the job queue is full it blocks until a
// worker frees a slot. Callers should therefore stop submitting before calling
// Shutdown. The send happens while holding a read lock, which guarantees the
// queue cannot be closed mid-send (the panic such a race would otherwise
// cause).
func (p *Pool[T]) Submit(job T) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return ErrClosed
	}
	p.jobs <- job
	return nil
}

// Shutdown stops the pool from accepting new jobs and waits for queued and
// in-flight jobs to drain, or until ctx is done — whichever happens first. It
// returns nil once the pool has drained, or ctx.Err() if the drain deadline was
// reached first.
//
// A non-nil return means goroutines may still be running the stragglers. To
// force them to stop, cancel the context that was passed to Start. Shutdown is
// idempotent and safe to call concurrently.
func (p *Pool[T]) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.jobs)
	}
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
