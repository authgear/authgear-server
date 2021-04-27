package apierrors

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Details errorutil.Details

type Cause interface{ Kind() string }

type StringCause string

func (c StringCause) Kind() string { return string(c) }
func (c StringCause) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
	}{string(c)})
}

type MapCause struct {
	CauseKind string
	Data      map[string]interface{}
}

func (c MapCause) Kind() string { return c.CauseKind }
func (c MapCause) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})
	data["kind"] = c.CauseKind
	for k, v := range c.Data {
		data[k] = v
	}
	return json.Marshal(data)
}

type APIError struct {
	Kind
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Info    map[string]interface{} `json:"info,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *APIError) HasCause(kind string) bool {
	if c, ok := e.Info["cause"].(Cause); ok {
		return c.Kind() == kind
	} else if cs, ok := e.Info["causes"].([]Cause); ok {
		for _, c := range cs {
			if c.Kind() == kind {
				return true
			}
		}
	}

	return false
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

func (k Kind) NewWithCause(msg string, c Cause) error {
	return k.NewWithInfo(msg, Details{"cause": c})
}

func (k Kind) NewWithCauses(msg string, cs []Cause) error {
	return k.NewWithInfo(msg, Details{"causes": cs})
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
func (e *skyerr) FillDetails(d errorutil.Details) {
	for key, value := range e.details {
		d[key] = value
	}
}

func IsAPIError(err error) bool {
	if _, ok := err.(*APIError); ok {
		return true
	}

	var e *skyerr
	if errors.As(err, &e) {
		return true
	}

	var v *validation.AggregatedError
	return errors.As(err, &v)
}

func AsAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError
	}

	var e *skyerr
	if errors.As(err, &e) {
		details := errorutil.CollectDetails(err, nil)
		return &APIError{
			Kind:    e.kind,
			Message: e.Error(),
			Code:    e.kind.Name.HTTPStatus(),
			Info:    errorutil.FilterDetails(details, APIErrorDetail),
		}
	}

	var v *validation.AggregatedError
	if errors.As(err, &v) {
		causes := make([]Cause, len(v.Errors))
		for i, c := range v.Errors {
			c := c
			causes[i] = &c
		}
		return &APIError{
			Kind:    ValidationFailed,
			Message: v.Message,
			Code:    ValidationFailed.Name.HTTPStatus(),
			Info: map[string]interface{}{
				"causes": causes,
			},
		}
	}

	return &APIError{
		Kind:    Kind{InternalError, "UnexpectedError"},
		Message: "unexpected error occurred",
		Code:    InternalError.HTTPStatus(),
	}
}

func IsKind(err error, kind Kind) bool {
	e := AsAPIError(err)
	return e != nil && e.Kind == kind
}

func NewBadRequest(msg string) error {
	return BadRequest.WithReason(string(BadRequest)).New(msg)
}

func NewInvalid(msg string) error {
	return Invalid.WithReason(string(Invalid)).New(msg)
}

func NewForbidden(msg string) error {
	return Forbidden.WithReason(string(Forbidden)).New(msg)
}

func NewInternalError(msg string) error {
	return InternalError.WithReason(string(InternalError)).New(msg)
}

func NewNotFound(msg string) error {
	return NotFound.WithReason(string(NotFound)).New(msg)
}

func NewDataRace(msg string) error {
	return DataRace.WithReason(string(DataRace)).New(msg)
}
