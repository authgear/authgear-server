package interaction

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrIncompatibleInput = errors.New("incompatible input type for this node")
var ErrSameNode = errors.New("the edge points to the same current node")

type ErrInputRequired struct {
	Inner error
}

func (e *ErrInputRequired) Error() string {
	return fmt.Sprintf("new input is required: %v", e.Inner)
}

func (e *ErrInputRequired) Unwrap() error {
	return e.Inner
}

type ErrClearCookie struct {
	Cookies []*http.Cookie
	Inner   error
}

func (e *ErrClearCookie) Error() string {
	return fmt.Sprintf("invalid cookie: %v", e.Inner)
}

func (e *ErrClearCookie) Unwrap() error {
	return e.Inner
}
