# Logging

A lightweight, structured **JSON logging** library built on top of
[logrus](https://github.com/sirupsen/logrus) and dedicated to the SkorLife
foundation logging standard.

## Features

- Structured JSON output ready for log aggregators.
- Five severity levels: `Critical`, `Error`, `Warning`, `Info`, `Debug`.
- Automatic enrichment: app name, host, caller file/line/method.
- Functional options to attach identity, references, trace errors, stack
  traces and custom fields.
- Configurable minimum log level at runtime.

## Installation

To install run the following command inside your project directory

```sh
go get github.com/go-production-grade/logging
```

## Vendoring to your project (recommended for all projects)

```sh
git clone ssh://git@github.com/go-production-grade.git
```

## Usage Example

### Logger Initialization

Call `Init` once at application start-up with your service name. Options such
as `WithPrettyPrint` are applied here.

```go
package main

import (
    "bitbucket.org/skor-skortech/st-kit/logging"
)

func main() {
    logging.Init("sk-risk-module")

    // Optionally control the minimum level emitted (defaults to INFO).
    // Levels: CRITICAL, ERROR, WARNING, INFO, DEBUG.
    _ = logging.SetLevel("DEBUG")

    logging.Info(
        "SUCCESS.INIT",
        "service started",
        logging.WithAppVersion(100),
    )
}
```

### Logging an Error

Each level (`Critical`, `Error`, `Warning`, `Info`, `Debug`) takes a `code`, a
`message`, and any number of options to enrich the structured log entry.

```go
logging.Critical(
    "ERR.SK.CONNECTION",
    "error fetching abc from xyz",
    logging.WithIdentity("SL000141311"),
    logging.WithReference(map[string]any{
        "host": "127.0.0.1",
    }),
    logging.WithTraceError(err),
    logging.WithStackTrace(true),
)
```

### Custom Fields

Use `WithField` to attach an ad-hoc key/value, or build your own option
helpers on top of it. Reserved fields set by the logger are never overwritten.

```go
func WithAppType(v string) logging.Option     { return logging.WithField("appType", v) }
func WithInstruction(v string) logging.Option { return logging.WithField("instruction", v) }

logging.Info(
    "ORDER.SETTLED",
    "settlement complete",
    WithAppType("merchant"),
    WithInstruction("settle"),
)
```

## Options

| Option                          | Description                                                       |
| ------------------------------- | ----------------------------------------------------------------- |
| `WithPrettyPrint(bool)`         | Pretty-print JSON instead of single-line output. Set in `Init`.   |
| `WithCallerLevel(int)`          | Adjust how many stack frames to skip when resolving the caller.   |
| `WithAppVersion(int)`           | Application version code.                                         |
| `WithIdentity(string)`          | User / customer identifier.                                       |
| `WithService(string)`           | Logical service name for the entry.                               |
| `WithReference(map[string]any)` | Arbitrary contextual reference data.                              |
| `WithReferenceID(string)`       | Correlation / reference ID.                                       |
| `WithTraceError(error)`         | Attach an error for back-tracing.                                 |
| `WithStackTrace(bool)`          | Include the current goroutine stack trace.                        |
| `WithDeviceID(string)`          | Originating device ID.                                            |
| `WithPlatform(string)`          | Originating platform (`ANDROID`, `IOS`).                          |
| `WithIP(string)`                | Client IP address.                                                |
| `WithSkorTransID(string)`       | Skorcard transaction ID.                                          |
| `WithSessionID(string)`         | Session identifier.                                               |
| `WithField(key, value)`         | Attach a single custom field (building block for custom options). |
| `WithSkip(bool)`                | Skip emitting the log entry.                                      |

> **Note:** `Critical` logs at logrus `FatalLevel` and always includes a stack
> trace.

## Sample Output

```json
{
  "app": "sk-risk-module",
  "code": "ERR.SK.CONNECTION",
  "file": "/app/main.go",
  "host": "host-01",
  "identity": "SL000141311",
  "level": "fatal",
  "line": 42,
  "method": "main.main",
  "msg": "error fetching abc from xyz",
  "ref": "{\"host\":\"127.0.0.1\"}",
  "service": "APP",
  "time": "2026-06-21T10:00:00Z"
}
```

## Logging Philosophy

The sections below summarise the standard every service is expected to follow.

### Parameter Description

| Parameter           | Description                                                                      |
| ------------------- | -------------------------------------------------------------------------------- |
| `host`              | The host name or IP address where the source application is hosted.              |
| `app`               | The application ID which is sending the log.                                     |
| `file`              | The code file name where the error originated.                                   |
| `method`            | The method name where the error originated.                                      |
| `line`              | The line number where the error originated.                                      |
| `type`              | The type / severity of the log.                                                  |
| `identity`          | Unique identity (e.g. user ID).                                                  |
| `service`           | Type of the service — Application, PostgresDB, REDIS, CBI, AWS, etc.             |
| `code`              | The error code.                                                                  |
| `desc`              | The error description.                                                           |
| `doc`               | Link to documentation describing the error code, its cause and resolution steps. |
| `ts`                | The time of the error in UTC.                                                    |
| `reference.request` | The entire request object in JSON format (request payload).                      |

### Log Levels

Listed in order of decreasing severity.

| Level        | Description                                                                                                                                                     | Alerts                  | Example                                                                                          |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------ |
| **Critical** | Application is broken; the last log before shutdown. Covers initialisation errors that prevent further operations.                                              | Enabled                 | Postgres not connected at app start.                                                             |
| **Warning**  | A breakdown was prevented by a fallback mechanism. Repeated warnings may lead to a fatal error. Needs supervision.                                              | Enabled                 | Mongo connection dropped, but retried and re-established without failure.                        |
| **Error**    | The current operation failed due to a critical error and similar future operations may also fail. Repeated occurrence beyond a threshold qualifies as critical. | Enabled, based on rules | Required data not found in PostgresDB; external service returns failure 150 times in 60 minutes. |
| **Info**     | Informational, always app-level logs.                                                                                                                           | Enabled, based on rules | App initialised successfully; Mongo connected successfully.                                      |
| **Debug**    | Debugging logs showing intermediate states of an operation.                                                                                                     | Disabled                | Computed payable amount before/after internet handling charges.                                  |

### Components

The originating point of an error:

1. **Datastore** — errors at SQL/NoSQL databases, in-memory data stores, etc.
2. **Queue** — errors at queueing technologies (RabbitMQ, MSMQ, Kafka, etc.).
3. **InternalServices** — errors at services hosted within the SkorLife domain.
4. **ExternalServices** — errors at services outside the SkorLife domain (CLIK, AKSHATA, INFOBIP, etc.).
5. **Application** — errors anywhere inside the application not covered above.

### Exception Log Categories

Applicable to `Critical`, `Warning` and `Error` logs.

| Category            | Description                                                                      |
| ------------------- | -------------------------------------------------------------------------------- |
| `ConnectionError`   | Destination endpoint not reachable (incorrect host/IP or endpoint down).         |
| `TimeoutError`      | Host took too long to reply, resulting in a connection timeout.                  |
| `DataNotFoundError` | Data missing from the lookup.                                                    |
| `ParseError`        | Failure while parsing serialised data.                                           |
| `ValidationError`   | Handled exception during request data validation.                                |
| `UnknownError`      | Caught at the global level (unhandled exception); category cannot be identified. |

### Log Transport

- The application should **not** open a connection to the log database.
- The application should log to standard output (`fmt.Println` / `console.log`).
- Logger agents read from standard output and forward logs to the log database.
- The channel between application and logger agent must be secured with SSL.

### What Should Be Logged

- Logs of all types should include the request object for request accountability.
- The application layer is responsible for masking sensitive data (KTP, mobile
  number, email, PCI data, etc.).
- `Critical` logs must always be emitted in **sync mode**.

### Mandatory Parameters

```jsonc
{
  "host": "",            // Host name of the running application (IP / Container ID)
  "app": "",             // Application ID identifying the running application

  "file": "",            // File name of the originating error
  "method": "",          // Method name of the originating error
  "line": "",            // Line number of the originating error

  "type": "",            // Type of log (Critical, Warning, Error, Info, Debug)
  "userID": "SL001",     // Unique identity
  "component": "",       // Originating point of the error

  "code": "",            // Error code
  "desc": "",            // Error description
  "category": "",        // Error category (ConnectionError, DataNotFoundError, etc.)
  "doc": "",             // Link to the documentation of the error

  "ts": "",              // Time the error occurred (UTC: dd MMM, yyyy hh:mm:ss)

  "reference": {         // Reference object (app-specific params may be added)
    "request": {}        // Request object
  }
}
```
