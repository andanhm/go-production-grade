package errors

// Option is a function adapter to change config of the errors struct.
type Option func(*option)

type option struct {
	message  string
	state    string
	errors   []error
	identity string
	service  Service
	data     any
}

// WithMessage message is a human-readable ASCII text providing additional information, used
// to assist the client developer in understanding the error that occurred.
func WithMessage(message string) func(*option) {
	return func(o *option) {
		o.message = message
	}
}

// WithData additional details need to handling the error
func WithData(data any) func(*option) {
	return func(o *option) {
		o.data = data
	}
}

// WithState its to determine the user journey state
// (Consider it as as App Activity OCR -> KYC) In between which state error occurred
func WithState(state string) func(*option) {
	return func(o *option) {
		o.state = state
	}
}

// WithTraceError keeps the original error for back tracing
func WithTraceError(err error) func(*option) {
	return func(o *option) {
		o.errors = append(o.errors, err)
	}
}

// WithIdentity a unique id for tracing error who got impacted
func WithIdentity(identity string) func(*option) {
	return func(o *option) {
		o.identity = identity
	}
}

// WithService determines which service throw the error
// mainly used to determine application level / 3rd party
func WithService(service Service) func(*option) {
	return func(o *option) {
		o.service = service
	}
}
