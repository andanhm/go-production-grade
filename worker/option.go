package worker

// Option is a function adapter to change the configuration of a Pool.
type Option func(*option)

type option struct {
	buffer int
}

// WithBuffer sets the capacity of the job queue. A buffer of 0 makes Submit
// block until a worker is ready (a synchronous hand-off). When unset, the
// buffer defaults to the pool size.
func WithBuffer(n int) Option {
	return func(o *option) {
		o.buffer = n
	}
}
