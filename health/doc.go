// Package health provides a small registry for probing the liveness of an
// application's dependencies (databases, caches, downstream services) and
// aggregating them into a single status report.
//
// Checks run concurrently, each bounded by a per-check timeout, so the report
// returns promptly even when a dependency is slow or unresponsive. The overall
// status is "fail" if any registered check fails, otherwise "ok".
package health
