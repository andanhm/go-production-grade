package errors

import (
	"context"
	"errors"
	"net"
)

// Is returns whether err informs of a Error if the provided type
// Error interface it asserts to the error build-in interface to Error
func Is(err error) (Error, bool) {
	if e, ok := err.(Error); ok {
		return e, ok
	}
	return nil, false
}

// As returns the AppError from the error chain if it exists.
// Replaces your "Unwrap" method.
func As(err error) (*appError, bool) {
	var e *appError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// IsTimeout checks if the error is a timeout error
// it checks if the error is a context deadline exceeded or a net.Error with Timeout() method
// if the error is a context deadline exceeded or a net.Error with Timeout() method it returns true
// otherwise it returns false
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var nErr net.Error
	if errors.As(err, &nErr) && nErr.Timeout() {
		return true
	}

	return false
}
