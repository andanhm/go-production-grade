package health

import "time"

// Option is a function adapter to change the configuration of a Health
// registry.
type Option func(*option)

type option struct {
	timeout time.Duration
}

// WithTimeout sets the maximum duration any single check may run before it is
// reported as failed. Defaults to 2s.
func WithTimeout(d time.Duration) Option {
	return func(o *option) {
		if d > 0 {
			o.timeout = d
		}
	}
}
