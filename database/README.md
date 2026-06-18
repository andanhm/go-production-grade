# Database Logging & Tracing

Example: Package provides on example of centralized error capturing and query tracing for both **Postgres (SQL via GORM)** and **MongoDB (NoSQL via MongoDB Go Driver)**.

By centralizing all database logs, we avoid repeating `if err != nil { log.Error(...) }` at every query site, enforce a consistent structure, and easily manage log levels or filter noise in one place.

## 1. Postgres & GORM Centralized Logging

GORM logs all operations through a simple `Printf(format string, args ...any)` interface. We route this through a custom choke point to classify, enrich, and filter messages.

### Postgres Key Components

- **Custom Logger Implementation**: [logger.go](./postgress/logger.go) implements the GORM logging writer interface.
- **Database Connection Setup**: [database.go](./postgress/database.go) registers this logger when initializing the Postgres connection.

### How Postgres Trace & Error Logging Works

1. **Production Mode (Errors Only)**: By default, the log level is set to `logger.Error` and `IgnoreRecordNotFoundError` is enabled. Only failed queries/database connection issues will reach `Printf`.
2. **Development Mode (Trace Mode)**: If `config.EnableLogging()` is true, the log level shifts to `logger.Info`. This enables SQL trace logging where every statement is logged.
3. **Filtering & Noise Control**: In [logger.go](./postgress/logger.go), common expected database errors like unique constraint violations (`23505`) are ignored.
4. **Context Cancellations**: Context cancellations and timeouts are dynamically demoted to warnings so they don't trigger false alarms.
5. **Business Identity Enrichment**: The logger uses regex to parse out identity tokens (like `SC123456`) from raw statements to correlate SQL errors with specific users/sessions.

```go
// Example configuration in postgress/database.go
logging := logger.New(&Logger{}, logger.Config{
    LogLevel:                  logger.Error, // Errors only in production
    IgnoreRecordNotFoundError: true,         // Avoid record-not-found log noise
    Colorful:                  false,
    SlowThreshold:             time.Second * 30,
})

if config.EnableLogging() {
    logging = logger.Default.LogMode(logger.Info) // Trace all queries in dev
}
```

## 2. MongoDB Centralized Logging

For MongoDB, we use the Go Driver's logger features to attach a custom log sink to the client.

### MongoDB Key Components

- **Logger Sink**: [logger.go](file:///Users/andanhm/go/src/github.com/go-production-grade/database/mongo/logger.go) implements the MongoDB driver log sink with `Info` and `Error` methods.
- **Client Configuration**: [mongo.go](file:///Users/andanhm/go/src/github.com/go-production-grade/database/mongo/mongo.go) hooks this sink into client setup.

### How Mongo Trace & Error Logging Works

1. **Trace Logs (`Info`)**:
   - Trace logs represent MongoDB command executions, connection events, and driver status.
   - If the environment variable `TIER` is set to `"production"`, info/trace logs are ignored.
   - For local development, trace logs write command detail maps into structured logs with key/value pairs using the built-in `log.Printf`.
2. **Error Logs (`Error`)**:
   - Emits driver failures and client query errors.
   - Automatically checks if an error represents client timeout or cancellation using `isContextCancellation` and downgrades it to warning prints.
   - Otherwise, errors are logged with the database query context.

```go
// Example client configuration in mongo/mongo.go
client, err := mongo.Connect(
    context.Background(),
    options.Client().SetConnectTimeout(time.Second*10),
    options.Client().
        ApplyURI(uri).
        SetLoggerOptions(options.Logger().
            SetComponentLevel(options.LogComponentConnection, options.LogLevelInfo).
            SetSink(&Logger{}), // Injecting the custom log sink
        ),
)
```

## Centralized Policies Applied

Both loggers enforce these design choices automatically:

- **Correlation Tracing**: Identifies business identities matching `\b(SC[0-9]{6,})\b` and extracts them as an independent field.
- **Severity Management**: Differentiates system errors from client-induced terminations (`context canceled`, `context deadline exceeded`, `timeout`).
- **No boilerplate at Call Sites**: Repository methods simply return errors up the stack, confident that logging and tracing are already taken care of

## Example JSON Logs

Since `logrus.JSONFormatter` is configured, all logs are structured in JSON format.

### 1. Postgres GORM Query Error

```json
{
  "code": "ERR.DB.QUERY",
  "fields.level": "error",
  "identity": "SC012202601",
  "level": "error",
  "msg": "Slow Query INSERT INTO \"SC_USER (user_id) VALUES ('SC012202601)\"",
  "ref": "{\"module\":\"gorm\",\"msg\":\"Slow Query INSERT INTO \\\"SC_USER (user_id) VALUES ('SC012202601)\\\"\",\"raw\":[\"INSERT INTO \\\"SC_USER (user_id) VALUES ('SC012202601)\\\"\"],\"result\":[\"INSERT INTO \\\"SC_USER (user_id) VALUES ('SC012202601)\\\"\"],\"type\":\"sql\"}",
  "service": "postgres",
  "time": "2026-06-18T15:48:11+05:30"
}
```

### 2. MongoDB Info/Trace Query

```json
{
  "code": "DB.QUERY",
  "fields.level": "info",
  "identity": "SC012202601",
  "level": "info",
  "msg": "Command find executed",
  "ref": "{\"data\":{\"collection\":\"users\",\"command\":\"find\",\"identity\":\"SC012202601\"},\"level\":1,\"module\":\"mongo\",\"type\":\"nosql\"}",
  "service": "mongodb",
  "time": "2026-06-18T15:59:44+05:30"
}
```

### 3. MongoDB Error Query

```json
{
  "code": "ERR.DB.QUERY",
  "fields.level": "error",
  "identity": "SC012202601",
  "level": "error",
  "msg": "Find query failed",
  "ref": "{\"data\":{\"identity\":\"SC012202601\"},\"error\":\"BadValue: query requires identity\",\"module\":\"mongo\",\"type\":\"nosql\"}",
  "service": "mongodb",
  "time": "2026-06-18T15:59:44+05:30"
}
```
