package transport

import (
	"net/url"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

func makeValidationError(fn func(*validation.Context)) error {
	ctx := &validation.Context{}
	fn(ctx)
	return ctx.Error("invalid parameters")
}

func getOptionalIntParam(q url.Values, name string) (*int, error) {
	s := q.Get(name)
	if s == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil, makeValidationError(func(ctx *validation.Context) {
			ctx.Child(name).EmitError("type", map[string]interface{}{"expected": "integer"})
		})
	}
	return &v, nil
}

func getIntParam(q url.Values, name string) (int, error) {
	v, err := getOptionalIntParam(q, name)
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, makeValidationError(func(ctx *validation.Context) {
			ctx.Child(name).EmitError("required", nil)
		})
	}
	return *v, nil
}

func getOptionalDateParam(q url.Values, name string) (*string, error) {
	s := q.Get(name)
	if s == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return nil, makeValidationError(func(ctx *validation.Context) {
			ctx.Child(name).EmitError("format", map[string]interface{}{"expected": "date"})
		})
	}
	return &s, nil
}

func getDateParam(q url.Values, name string) (string, error) {
	v, err := getOptionalDateParam(q, name)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", makeValidationError(func(ctx *validation.Context) {
			ctx.Child(name).EmitError("required", nil)
		})
	}
	return *v, nil
}

func validateMonth(name string, v int) error {
	if v < 1 || v > 12 {
		return makeValidationError(func(ctx *validation.Context) {
			ctx.Child(name).EmitError("range", map[string]interface{}{"minimum": 1, "maximum": 12})
		})
	}
	return nil
}
