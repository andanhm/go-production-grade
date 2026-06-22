// Package signals provides clean interception of OS termination signals
// (SIGINT, SIGTERM) for graceful application shutdown.
//
// It exposes a single primitive, Context, which derives a context that is
// cancelled when a termination signal is received. Wiring shutdown logic to a
// context — rather than a bare signal channel — lets the same cancellation
// propagate to HTTP servers, background workers and database clients that
// already accept a context.Context.
package signals
