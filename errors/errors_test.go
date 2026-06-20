package errors_test

import (
	"fmt"
	"testing"

	"github.com/go-production-grade/errors"
	"github.com/stretchr/testify/assert"
)

func Example_new() {
	err := errors.New(
		"AG001",
		"Mobile number not found in the system",
		errors.WithMessage("Input error"),
	)
	if err, ok := errors.As(err); ok {
		fmt.Printf("UI Error %v", err.Message)
		return
	}
	if err, ok := errors.Is(err); ok {
		fmt.Printf("Error code %v", err.Code())
		return
	}
	fmt.Printf("Error %v", err.Error())
}

func Test_New(t *testing.T) {
	err := errors.New(
		"AG001",
		"Mobile number not found in the system",
		errors.WithMessage("Input error"),
	)
	assert.Error(t, err)
	assert.Equal(t, "AG001", err.Code())
	assert.Equal(t, "Mobile number not found in the system", err.Error())
	errUnwrap, ok := errors.As(err)
	if ok {
		assert.Equal(t, "Input error", errUnwrap.Message)
	}
}
