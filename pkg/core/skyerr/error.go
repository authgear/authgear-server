package skyerr

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type Details errors.Details

type APIError struct {
	Kind
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Info    map[string]interface{} `json:"info,omitempty"`
}

func (k Kind) New(msg string) error {
	return &skyerr{kind: k, msg: msg}
}

func (k Kind) NewWithDetails(msg string, details Details) error {
	return &skyerr{kind: k, msg: msg, details: details}
}

func (k Kind) NewWithInfo(msg string, info Details) error {
	d := Details{}
	for k, v := range info {
		d[k] = APIErrorDetail.Value(v)
	}
	return k.NewWithDetails(msg, d)
}

func (k Kind) Wrap(err error, msg string) error {
	return &skyerr{kind: k, inner: err, msg: msg}
}

func (k Kind) Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	return k.Wrap(err, err.Error())
}

type skyerr struct {
	inner   error
	msg     string
	kind    Kind
	details Details
}

func (e *skyerr) Error() string { return e.msg }
func (e *skyerr) Unwrap() error { return e.inner }
func (e *skyerr) FillDetails(d errors.Details) {
	for key, value := range e.details {
		d[key] = value
	}
}

func AsAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	var e *skyerr
	if !errors.As(err, &e) {
		return &APIError{
			Kind:    Kind{InternalError, "UnexpectedError"},
			Message: "unexpected error occurred",
			Code:    InternalError.HTTPStatus(),
		}
	}

	details := errors.CollectDetails(err, nil)
	return &APIError{
		Kind:    e.kind,
		Message: e.Error(),
		Code:    e.kind.Name.HTTPStatus(),
		Info:    errors.FilterDetails(details, APIErrorDetail),
	}
}

func IsKind(err error, kind Kind) bool {
	var e *skyerr
	if !errors.As(err, &e) {
		return false
	}
	return e.kind == kind
}

func NewBadRequest(msg string) error {
	return BadRequest.WithReason(string(BadRequest)).New(msg)
}

func NewInvalid(msg string) error {
	return Invalid.WithReason(string(Invalid)).New(msg)
}

func NewInternalError(msg string) error {
	return InternalError.WithReason(string(InternalError)).New(msg)
}

func NewNotFound(msg string) error {
	return NotFound.WithReason(string(NotFound)).New(msg)
}
