# Health

A small registry for probing the liveness of an application's dependencies
(databases, caches, downstream services) and aggregating them into a single
status report. It is the engine behind the [`/health`](../api) endpoint.

## Features

- Register a named `CheckFunc` per dependency.
- Checks run **concurrently**, each bounded by a per-check timeout, so the
  report returns promptly even when a dependency is slow or wedged.
- Honours the caller's context — a cancelled request never hangs the report.
- Overall status is `fail` if any check fails, otherwise `ok` (matching the
  health-endpoint contract).
- Per-check latency and error are captured. Zero third-party dependencies.

## Installation

```sh
go get github.com/go-production-grade/health
```

## Usage

```go
checks := health.New(health.WithTimeout(2 * time.Second))

checks.Register("postgres", func(ctx context.Context) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    return sqlDB.PingContext(ctx)
})

checks.Register("mongo", func(ctx context.Context) error {
    return mongoDB.Client().Ping(ctx, readpref.Primary())
})

report := checks.Check(ctx)
// report.Status        -> "ok" | "fail"
// report.Checks["mongo"] -> {Status, Error, Latency}
```

## API

| Symbol                            | Description                                                                                  |
| --------------------------------- | -------------------------------------------------------------------------------------------- |
| `health.New(opts...)`             | Create a registry. `WithTimeout(d)` bounds each check (default 2s).                          |
| `(*Health).Register(name, fn)`    | Add a named check. `fn` must honour its context.                                             |
| `(*Health).Check(ctx) Report`     | Run all checks concurrently and aggregate. Returns promptly; never hangs on a stuck probe.   |
| `health.StatusOK` / `StatusFail`  | The two status values (`"ok"` / `"fail"`).                                                   |

### Report shape

```jsonc
{
  "status": "fail",            // "ok" if every check passed
  "checks": {
    "postgres": { "status": "ok",   "latency": "1.2ms" },
    "mongo":    { "status": "fail", "error": "context deadline exceeded", "latency": "2s" }
  }
}
```

## Design note

`Check` spawns one goroutine per registered check into a buffered channel and
collects results under a `select` on `ctx.Done()`. If the caller's deadline
fires before all checks settle, the unsettled ones are reported as failed and
the report returns immediately — the same "return promptly, don't hang"
discipline as `worker.Shutdown`.
