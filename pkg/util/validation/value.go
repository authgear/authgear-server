package validation

import (
	"context"
	"errors"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type Validator interface {
	Validate(context.Context, *Context)
}

type Context struct {
	pointer jsonpointer.T
	errors  *[]Error
}

func (c *Context) Child(path ...string) *Context {
	if c.errors == nil {
		c.errors = &[]Error{}
	}
	return &Context{pointer: append(c.pointer, path...), errors: c.errors}
}

func (c *Context) EmitError(keyword string, info map[string]interface{}) {
	if c.errors == nil {
		c.errors = &[]Error{}
	}
	*c.errors = append(*c.errors, Error{Location: c.pointer.String(), Keyword: keyword, Info: info})
}

func (c *Context) EmitErrorMessage(msg string) {
	c.EmitError("general", map[string]interface{}{"msg": msg})
}

func (c *Context) AddError(err error) {
	if err == nil {
		return
	}

	var aErr *AggregatedError
	if errors.As(err, &aErr) {
		if aErr == nil || len(aErr.Errors) == 0 {
			return
		}
		if c.errors == nil {
			c.errors = &[]Error{}
		}
		for _, err := range aErr.Errors {
			err.Location = c.pointer.String() + err.Location
			*c.errors = append(*c.errors, err)
		}
	} else {
		c.EmitErrorMessage(err.Error())
	}
}

func (c *Context) Validate(ctx context.Context, value interface{}) {
	if v, ok := value.(Validator); ok {
		v.Validate(ctx, c)
	}
}

func (c *Context) Error(msg string) error {
	if c.errors == nil || len(*c.errors) == 0 {
		return nil
	}
	return &AggregatedError{Message: msg, Errors: *c.errors}
}

func ValidateValue(ctx context.Context, value interface{}) error {
	return ValidateValueWithMessage(ctx, value, defaultErrorMessage)
}

func ValidateValueWithMessage(ctx context.Context, value interface{}, msg string) error {
	valicationCtx := &Context{}
	valicationCtx.Validate(ctx, value)
	return valicationCtx.Error(msg)
}
