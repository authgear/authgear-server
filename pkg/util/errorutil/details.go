package errorutil

import (
	"errors"
)

type Details map[string]interface{}

type Detailer interface {
	error
	FillDetails(d Details)
}

type errorDetails struct {
	inner   error
	details Details
}

func WithDetails(err error, d Details) error {
	return &errorDetails{err, d}
}

func (e *errorDetails) Error() string { return e.inner.Error() }
func (e *errorDetails) Unwrap() error { return e.inner }
func (e *errorDetails) FillDetails(d Details) {
	for key, value := range e.details {
		d[key] = value
	}
}

func CollectDetails(err error, d Details) Details {
	if d == nil {
		d = Details{}
	}

	// Inspect the error chain to fill out Detailer.
	var detailers []Detailer
	for err != nil {
		var detailer Detailer
		ok := errors.As(err, &detailer)
		if ok {
			detailers = append(detailers, detailer)
		}
		err = errors.Unwrap(err)
	}

	// Loop the detailers backward to make sure wrapping error override wrapped error.
	for i := len(detailers) - 1; i >= 0; i-- {
		detailer := detailers[i]
		detailer.FillDetails(d)
	}

	return d
}
