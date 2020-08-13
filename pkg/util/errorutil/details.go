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
	var detailer Detailer
	for errors.As(err, &detailer) {
		detailer.FillDetails(d)
		err = errors.Unwrap(detailer)
	}
	return d
}
