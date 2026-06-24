package health

import (
	"context"
	"sync"
	"time"
)

// Status is the health status of the system or a single dependency.
type Status string

const (
	// StatusOK indicates the system / dependency is healthy.
	StatusOK Status = "ok"
	// StatusFail indicates the system / dependency is unhealthy.
	StatusFail Status = "fail"
)

// defaultTimeout bounds each individual check when none is configured.
const defaultTimeout = 2 * time.Second

// CheckFunc probes a single dependency. It must honour ctx so a slow or stuck
// dependency cannot wedge the overall report. A nil error means healthy.
type CheckFunc func(ctx context.Context) error

// Check is the outcome of probing one dependency.
type Check struct {
	// Status is "ok" or "fail".
	Status Status `json:"status"`
	// Error is the failure reason, omitted when healthy.
	Error string `json:"error,omitempty"`
	// Latency is the human-readable probe duration (e.g. "1.2ms").
	Latency string `json:"latency"`
}

// Report is the aggregated result of running every registered check.
type Report struct {
	// Status is the overall status: "fail" if any check failed, else "ok".
	Status Status `json:"status"`
	// Checks holds the per-dependency results keyed by registered name.
	Checks map[string]Check `json:"checks"`
}

type checker struct {
	name string
	fn   CheckFunc
}

// Health is a registry of named dependency checks.
type Health struct {
	timeout  time.Duration
	mu       sync.RWMutex
	checkers []checker
}

// New creates a registry. By default each check is bounded to 2s; override with
// WithTimeout.
func New(opts ...Option) *Health {
	opt := option{timeout: defaultTimeout}
	for _, o := range opts {
		o(&opt)
	}
	return &Health{timeout: opt.timeout}
}

// Register adds a named check. Registering the same name twice keeps both; use
// distinct names per dependency.
func (h *Health) Register(name string, fn CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker{name: name, fn: fn})
}

// Check runs every registered check concurrently, each bounded by the
// configured per-check timeout (and by ctx). It returns once all checks settle
// or ctx is done — it never blocks longer than the slowest of those two, so a
// misbehaving probe cannot hang the caller. The overall status is "fail" if any
// check failed.
func (h *Health) Check(ctx context.Context) Report {
	h.mu.RLock()
	checkers := make([]checker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

	report := Report{Status: StatusOK, Checks: make(map[string]Check, len(checkers))}

	type result struct {
		name  string
		check Check
	}
	results := make(chan result, len(checkers))

	for _, c := range checkers {
		go func(c checker) {
			results <- result{name: c.name, check: h.run(ctx, c.fn)}
		}(c)
	}

	for range checkers {
		select {
		case r := <-results:
			report.Checks[r.name] = r.check
			if r.check.Status == StatusFail {
				report.Status = StatusFail
			}
		case <-ctx.Done():
			// The caller's deadline fired before all checks settled. Report
			// what we have plus the unsettled ones as failed, and return.
			report.Status = StatusFail
			for _, c := range checkers {
				if _, done := report.Checks[c.name]; !done {
					report.Checks[c.name] = Check{Status: StatusFail, Error: ctx.Err().Error()}
				}
			}
			return report
		}
	}
	return report
}

// run probes a single dependency under a bounded context and records latency.
func (h *Health) run(ctx context.Context, fn CheckFunc) Check {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	start := time.Now()
	err := fn(ctx)
	latency := time.Since(start).String()

	if err != nil {
		return Check{Status: StatusFail, Error: err.Error(), Latency: latency}
	}
	return Check{Status: StatusOK, Latency: latency}
}
