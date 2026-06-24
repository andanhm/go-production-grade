// Package worker provides a bounded pool of background goroutines that drain
// gracefully on termination.
//
// It complements the signals package: a termination signal cancels a context,
// the application stops accepting new work, the worker pool drains in-flight
// and queued jobs within a bounded deadline, and only then are connection
// managers (databases, caches, queues) closed — in that order, because
// draining jobs may still depend on those connections.
package worker
