# Worker

A bounded pool of background goroutines that **drain gracefully** on
termination. It pairs with the [`signals`](../signals) package to give an
application a clean shutdown sequence: stop accepting work → drain in-flight
jobs → close connection managers.

## Features

- Generic `Pool[T]` — process jobs of any type with a fixed number of workers.
- Graceful drain modelled on `http.Server.Shutdown`: in-flight and queued jobs
  finish; new submissions are rejected.
- Separate work / drain contexts — the drain never aborts the jobs it is
  waiting on, and a forced abort is always one `cancel()` away.
- Bounded deadline: `Shutdown` returns promptly with `context.DeadlineExceeded`
  instead of hanging on a stuck handler.
- Backpressure via a configurable buffer; `Submit` blocks when the queue is
  full rather than growing memory unbounded.
- Race-clean (`go test -race`), zero third-party dependencies.

## The two-context model

This is the one thing to understand. A worker pool has **two** lifecycles, and
conflating them is the classic shutdown bug.

| Context              | Passed to    | Meaning                                                                    |
| -------------------- | ------------ | -------------------------------------------------------------------------- |
| **work context**     | `Start(ctx)` | Propagated to every handler. Cancelling it means *abort running work now*. |
| **drain context**    | `Shutdown(ctx)` | A deadline for how long to wait for queued + in-flight jobs to finish.  |

`Shutdown` closes the job queue (so workers exit once drained) but leaves the
**work context live**, so handlers already running can complete. If the drain
deadline is hit first, `Shutdown` returns `ctx.Err()` and the stragglers are
*still running* — cancel the work context to force them to stop.

```
graceful:  Shutdown(drainCtx) closes queue → handlers finish → returns nil
timed-out: Shutdown(drainCtx) hits deadline → returns err → caller cancels Start ctx → workers abort
```

## Installation

```sh
go get github.com/go-production-grade/worker
```

## Usage Example

### Basic drain

```go
package main

import (
    "context"
    "time"

    "github.com/go-production-grade/worker"
)

func main() {
    pool := worker.New(8, func(ctx context.Context, url string) {
        // ... process the job, honouring ctx for cancellation ...
    })

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    pool.Start(ctx)

    _ = pool.Submit("https://example.com/a")
    _ = pool.Submit("https://example.com/b")

    // Drain with a bounded deadline.
    drainCtx, drainCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer drainCancel()

    if err := pool.Shutdown(drainCtx); err != nil {
        // Drain timed out — force the stragglers to abort.
        cancel()
    }
}
```

### Full termination sequence (signals + workers + connection managers)

The ordering here is the whole point of "graceful drain of background workers
**and** connection managers": workers must be **fully drained before** any
connection is closed, because in-flight jobs may still be querying the database.
Do not close connections in sibling `defer`s — sequence them explicitly.

```go
package main

import (
    "context"
    "time"

    "github.com/go-production-grade/logging"
    "github.com/go-production-grade/signals"
    "github.com/go-production-grade/worker"

    "gorm.io/gorm"
)

func run(db *gorm.DB) {
    logging.Init("sk-worker")

    // 1. A context cancelled on SIGINT/SIGTERM (and the same ctx drives work).
    ctx, stop := signals.Context(context.Background())
    defer stop()

    // 2. Start the background workers.
    pool := worker.New(8, func(ctx context.Context, job Job) {
        process(ctx, db, job) // queries db using ctx
    })
    pool.Start(ctx)

    go produce(ctx, pool) // feed jobs from a queue/stream until ctx is done

    // 3. Block until a termination signal arrives.
    <-ctx.Done()
    logging.Info("APP.SHUTDOWN", "draining workers")

    // 4. Drain the pool with a bounded deadline. Use a FRESH context — ctx is
    //    already cancelled, and we want in-flight jobs to finish, not abort.
    drainCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    if err := pool.Shutdown(drainCtx); err != nil {
        logging.Error("ERR.WORKER.DRAIN", err.Error(), logging.WithTraceError(err))
    }

    // 5. Only NOW that no job is running do we close connection managers.
    if sqlDB, err := db.DB(); err == nil {
        if err := sqlDB.Close(); err != nil {
            logging.Error("ERR.DB.CLOSE", err.Error(), logging.WithTraceError(err))
        }
    }
    logging.Info("APP.STOPPED", "shutdown complete")
}
```

> **Why a fresh `drainCtx`?** The signal context (`ctx`) is already cancelled by
> the time you reach the drain step. Reusing it would make `Shutdown` return
> immediately and abort every in-flight job — the opposite of a graceful drain.

## API

| Symbol                              | Description                                                                                              |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `worker.New[T](size, handler, ...)` | Create a pool of `size` workers running `handler` per job. Size `< 1` is clamped to 1.                  |
| `(*Pool[T]).Start(ctx)`             | Launch the workers. Idempotent. `ctx` is the work context propagated to every handler.                  |
| `(*Pool[T]).Submit(job) error`      | Enqueue a job. Returns `ErrClosed` once shutting down. Blocks (backpressure) when the buffer is full.    |
| `(*Pool[T]).Shutdown(ctx) error`    | Stop accepting work and drain queued + in-flight jobs until `ctx` is done. Idempotent. Returns `ctx.Err()` on timeout. |
| `worker.WithBuffer(n)`              | Set the job-queue capacity (default = pool size; `0` = synchronous hand-off).                            |
| `worker.ErrClosed`                  | Returned by `Submit` after shutdown has begun.                                                           |

## Notes

- **Stop submitting before `Shutdown`.** `Submit` blocks under backpressure; if
  you keep submitting during a drain you may block on a full queue. Tie the
  producer's lifetime to the work context (see step 4 above).
- **`Submit` is safe against the close race.** The send happens under a read
  lock while `Shutdown` closes the queue under a write lock, so a "send on
  closed channel" panic cannot occur.
- A non-nil `Shutdown` error means goroutines may still be running — cancel the
  `Start` context to force them down before exiting the process.
