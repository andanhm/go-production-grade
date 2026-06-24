package api

// Option is a function adapter to change the configuration of the API.
type Option func(*option)

type option struct {
	version string
}

// WithVersion sets the build version reported by the health endpoint. Defaults
// to "v0.1.0".
func WithVersion(version string) Option {
	return func(o *option) {
		if version != "" {
			o.version = version
		}
	}
}
