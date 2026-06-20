# errors

Package errors provides an easy way to annotate errors without losing the
original error context.

An comprehensive guide line defined to followed by every developer

Usage:
This is particularly useful when you want to understand the state of execution when an error was returned from 3rd party authentication source.

It provides the type *appError which implements the standard golang error interface, so you can use this library interchangeably with code that is expecting a normal error return.

```bash
go get github.com/go-production-grade/errors
```

## Usage

This is particularly useful when you want to understand the state of execution when an error was returned from 3rd party authentication source.
It provides the type AppError which implements the standard golang error interface, so you can use this library interchangeably with code that is expecting a normal error return.

The traditional error handling idiom in Go

```go
if err != nil {
    return err
}
```

## Adding context to an error

The errors.New function returns a new error that adds context to the original error.

### Example

```go
err := authService.Authenticated("token")
if err != nil {
    // Note: AG<> is just reference code (Keep the application name) (Integrated Service [IS401])
    return errors.New(
            "AG401",
            err.Error(),
            errors.WithIdentity("SL20220601"),
            errors.WithTraceError(err),
        )
}
```

## Determine the cause of an error

```go
import "github.com/go-production-grade/errors"

// Determines the whether the provide token is valid or not.
// `access_token` could be user, app, refresh token
_, err = authService.Authenticated(ctx, &auth.Request{
    AccessToken: "<access_token>",
})
if err != nil {
    if err, ok := errors.Is(err); ok {
        switch code := err.Code(); code {
        case "AG002":
            {
                // Handle invalid request server could not understand the request due to invalid syntax.
                break
            }
        case "AG403":
            {
                // Handle user/app don't have authorized to access the feature (forbidden)
                break
            }
        case "AG401":
            {
                // error : invalid_token
                // The access token provided is expired, revoked, malformed, or invalid for other reasons.
                break
            }
        default:
            {
                // Handle other case
                log.Fatalf("Authenticated error: %v ", err)
                break
            }
        }
    }
}
```

## Addition context in an error

```go
if err, ok := errors.Is(err); ok {
    err.Code()
    err.Error()
    err.Date()
}
```

| Method     | Description                                                                                                                                               |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `code`     | Code is the error code and must have one of the values defines by the package                                                                             |
| `message`  | Description is a human-readable ASCII text providing additional information, used to assist the client developer in understanding the error that occurred |
| `identity` | Identify a user with a unique ID to track user activity                                                                                                   |
| `state`    | State if the request included the state parameter. Set to the value received from the Client. **(Not used)**                                              |
| `error`    | Error returns the underlying error's message.                                                                                                             |
| `errors`   | Trace of error message it always recommend to share with the application facing API                                                                       |

## Error Codes Examples

Below you will find a summary of error codes and corresponding HTTP status codes for every  flow Base supports.

| Code      |                       Message                       | HTTP Status Code |
| :-------- | :-------------------------------------------------: | ---------------: |
| **AG401** | Invalid / Expired token provided for authentication |              401 |
| **AG404** |     Indicates the user not found in our system      |              404 |
