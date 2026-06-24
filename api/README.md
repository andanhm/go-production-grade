# API

The application's HTTP surface. Currently exposes the **health diagnostics
endpoint**, which reports the overall application status and the status of every
dependency it relies on.

It is served from the gin server in [`main.go`](../main.go), wired for graceful
shutdown via the [`signals`](../signals) package, and backed by the
[`health`](../health) registry.

## The `/health` endpoint

```go
checks := health.New(health.WithTimeout(2 * time.Second))
checks.Register("self", func(context.Context) error { return nil })

app := api.New(checks, api.WithVersion("v0.1.0"))

v2 := router.Group("/v2")
v2.GET("/health", app.HealthHandler())
```

### Response

```jsonc
{
  "version": "v0.1.0",
  "host": "pod-7d9f",
  "commit": "abc123",          // from COMMIT_ID env
  "tier": "staging",           // from TIER env
  "status": "ok",              // "fail" if any dependency check fails
  "startedAt": "2026-06-23T23:02:25+05:30",
  "dependencies": {
    "postgres": { "status": "ok",   "latency": "1.2ms" },
    "redis":    { "status": "fail", "error": "connection refused", "latency": "2s" }
  }
}
```

### HTTP status code — readiness, not liveness

The handler returns **`200 OK`** when healthy and **`503 Service Unavailable`**
when any dependency check fails. This makes `/health` a **readiness** signal:
load balancers and orchestrators stop routing traffic to an instance that can't
reach its dependencies.

> Do **not** point a Kubernetes *liveness* probe at this endpoint. A liveness
> probe that 503s during a database outage will restart the pod — which can't
> fix the database — and cause a crash loop. Use it for readiness (and the LB),
> or split liveness onto a separate always-200 endpoint.

> Two intentional changes from the original sketch: the response is served with
> `ctx.JSON` (not `JSONP` — there's no cross-origin script-injection use case
> for a health endpoint), and the status code reflects health rather than always
> being `200`.

## Wiring real dependency checks

The server boots without a live database by registering only a `self` check.
Register a check per real connection as its handle becomes available — only when
the handle is non-nil, so the server stays runnable in environments without that
dependency:

```go
checks := health.New(health.WithTimeout(2 * time.Second))

if db != nil { // *gorm.DB
    checks.Register("postgres", func(ctx context.Context) error {
        sqlDB, err := db.DB()
        if err != nil {
            return err
        }
        return sqlDB.PingContext(ctx)
    })
}

if mongoDB != nil { // *mongo.Database
    checks.Register("mongo", func(ctx context.Context) error {
        return mongoDB.Client().Ping(ctx, readpref.Primary())
    })
}
```

## API

| Symbol                      | Description                                                                              |
| --------------------------- | ---------------------------------------------------------------------------------------- |
| `api.New(checks, opts...)`  | Build the API from a `*health.Health` registry. `WithVersion(v)` sets the build version. |
| `(*API).Health(ctx) Health` | Run all checks and assemble the report DTO.                                              |
| `(*API).HealthHandler()`    | gin handler: serves the report as JSON with a 200/503 status code.                       |
| `(Health).Healthy() bool`   | Whether the overall status is `"ok"`.                                                    |
