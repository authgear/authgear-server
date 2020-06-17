package validation

import (
	"errors"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type Validator interface {
	Validate(*Context)
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
		for _, err := range aErr.Errors {
			err.Location = c.pointer.String() + err.Location
			*c.errors = append(*c.errors, err)
		}
	} else {
		c.EmitErrorMessage(err.Error())
	}
}

func (c *Context) Validate(value interface{}) {
	if v, ok := value.(Validator); ok {
		v.Validate(c)
	}
}

func (c *Context) Error() error {
	if c.errors == nil {
		return nil
	}
	return convertErrors(*c.errors, nil)
}

func ValidateValue(value interface{}) error {
	ctx := &Context{}
	ctx.Validate(value)
	return ctx.Error()
}
