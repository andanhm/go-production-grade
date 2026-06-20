package errors

type (
	// Error is the interface providing the implementation needed to return error
	Error interface {
		// Code is the error code and must have one of the values defines by the package
		// constants.
		Code() string

		// Error returns the underlying error's message.
		Error() string

		// Data additional error details helps in passing between the packages
		Data() any
	}

	// agent2Error is a simple implementation of Error.
	appError struct {
		// Mandatory values always exposed
		ErrorCode string `json:"code"`
		Message   string `json:"message,omitempty"`

		ErrorDescription string  `json:"error,omitempty"`
		ErrorState       string  `json:"state,omitempty"`
		Service          Service `json:"service,omitempty"`
		Identity         string  `json:"identity,omitempty"`
		AdditionData     any     `json:"data,omitempty"`
		// Recommended to omit in case public facing api
		Errors []error `json:"errors,omitempty"`
	}
)

// New creates an error suitable to be returned in the application error responses.
func New(code string, description string, opts ...Option) Error {
	option := &option{}
	for _, opt := range opts {
		opt(option)
	}

	return &appError{
		ErrorCode:        code,
		ErrorDescription: description,
		Message:          option.message,
		ErrorState:       option.state,
		Errors:           option.errors,
		Service:          option.service,
		Identity:         option.identity,
		AdditionData:     option.data,
	}
}

func (e *appError) Code() string  { return e.ErrorCode }
func (e *appError) Data() any     { return e.AdditionData }
func (e *appError) Error() string { return e.ErrorDescription }
