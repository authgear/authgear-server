package errorutil

import (
	"errors"
	"log/slog"
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
	if err == nil {
		return nil
	}

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

	Unwrap(err, func(err error) {
		var detailer Detailer
		ok := errors.As(err, &detailer)
		if ok {
			detailers = append(detailers, detailer)
		}
	})

	// Loop the detailers backward to make sure wrapping error override wrapped error.
	for i := len(detailers) - 1; i >= 0; i-- {
		detailer := detailers[i]
		detailer.FillDetails(d)
	}

	return d
}

func (m Details) ToSlogAttrs() []slog.Attr {
	if m == nil {
		return nil
	}

	attrs := make([]slog.Attr, 0, len(m))

	for key, value := range m {
		attrs = append(attrs, slog.Any(key, value))
	}

	return attrs
}
