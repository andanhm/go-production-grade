# Signals

Clean interception of OS termination signals (`SIGINT`, `SIGTERM`) for graceful
application shutdown. The package exposes a single primitive that turns an
incoming signal into a cancelled `context.Context`, so the same shutdown event
propagates to every component that already speaks `context` — HTTP servers,
background workers, and database clients.

## Features

- Watches `SIGINT` **and** `SIGTERM` by default — the signal that actually
  matters in containers and orchestrators.
- Returns a `context.Context`, not a bare channel, so cancellation composes with
  `http.Server.Shutdown`, `gorm`, the official Mongo driver, etc.
- Automatic force-quit escape hatch: a second signal hits Go's default handler
  and kills the process, so a wedged shutdown can always be aborted.
- Honours parent-context cancellation as a shutdown trigger too.
- Zero third-party dependencies — standard library only.

## Why not just `signal.Notify`?

A very common bootstrap looks like this:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt) // SIGINT only!
<-quit
```

This works on a laptop but is subtly wrong in production:

| Problem                        | Impact                                                                                                                                                                                    |
| ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Only `os.Interrupt` (`SIGINT`) | Kubernetes / systemd / Docker send **`SIGTERM`** to stop a pod. SIGINT is never received, so the process is hard-killed with `SIGKILL` after the grace period and in-flight work is lost. |
| No second-signal escape hatch  | If shutdown hangs, the operator cannot interrupt it without `kill -9`.                                                                                                                    |
| A channel, not a context       | The signal cannot be propagated to `Shutdown(ctx)`, query contexts, or worker loops.                                                                                                      |

`signals.Context` fixes all three.

## Installation

```sh
go get github.com/go-production-grade/signals
```

## Usage Example

### Minimal

```go
package main

import (
    "context"

    "github.com/go-production-grade/signals"
)

func main() {
    // Cancelled on SIGINT or SIGTERM. stop() releases the handler.
    ctx, stop := signals.Context(context.Background())
    defer stop()

    // ... start your server / workers ...

    <-ctx.Done() // blocks until a termination signal arrives
    // ... run your graceful shutdown ...
}
```

### Graceful HTTP shutdown with Gin

This is the production wiring referenced by the snippet most projects start
from — corrected to listen for `SIGTERM`, drive shutdown from the context, and
check the `Shutdown` error.

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "github.com/go-production-grade/logging"
    "github.com/go-production-grade/signals"
)

func main() {
    logging.Init("sk-api")

    router := gin.New()
    router.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

    // Cancelled on SIGINT/SIGTERM. A second signal force-kills the process.
    ctx, stop := signals.Context(context.Background())
    defer stop()

    srv := &http.Server{
        Addr:    fmt.Sprintf("%s:%s", "0.0.0.0", "8080"),
        Handler: router,
    }

    // Serve in the background; ListenAndServe blocks until Shutdown is called.
    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            logging.Critical("ERR.SERVER.START", err.Error(), logging.WithTraceError(err))
        }
    }()
    logging.Info("APP.STARTED", "http server listening")

    // Block until a termination signal (or ctx cancellation) arrives.
    <-ctx.Done()
    logging.Info("APP.SHUTDOWN", "draining in-flight requests")

    // Give in-flight requests a bounded window to drain. Use a fresh context —
    // ctx is already cancelled.
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        logging.Error("ERR.SERVER.SHUTDOWN", err.Error(), logging.WithTraceError(err))
        return
    }
    logging.Info("APP.STOPPED", "shutdown complete")
}
```

> **Note:** `signals` only watches `SIGINT`/`SIGTERM`. `SIGKILL` (`kill -9`) and
> `SIGSTOP` cannot be caught by any program — design your grace period around
> the orchestrator's `terminationGracePeriodSeconds`.

### Watching custom signals

```go
// Reload-on-SIGHUP, for example, in addition to the defaults.
ctx, stop := signals.Context(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
defer stop()
```

## API

| Symbol                            | Description                                                                                                                                                  |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `signals.Default`                 | `[]os.Signal{SIGINT, SIGTERM}` — used when no signals are passed to `Context`.                                                                               |
| `signals.Context(parent, sig...)` | Returns a child context cancelled on the first matching signal or on parent cancellation, plus a `stop` (`context.CancelFunc`) that unregisters the handler. |

## How it works

`Context` is a thin, intention-revealing wrapper over the standard library's
[`signal.NotifyContext`](https://pkg.go.dev/os/signal#NotifyContext) with a
production-safe default signal set. The manual channel equivalent is:

```go
ch := make(chan os.Signal, 1)
signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
sig := <-ch          // wait
signal.Stop(ch)      // detach so a second signal force-quits
```

`Context` does exactly this, hands you a context instead of a channel, and wires
in the `signal.Stop` step so the second-signal force-quit behaviour is automatic.
