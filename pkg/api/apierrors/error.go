package apierrors

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/copyutil"
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

func (e *APIError) Clone() *APIError {
	ee, err := copyutil.Clone(e)
	if err != nil {
		return nil
	}
	return ee.(*APIError)
}

func (k Kind) New(msg string) error {
	return k.NewWithDetails(msg, make(Details))
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
	return errors.Join(&skyerr{kind: k, msg: msg}, err)
}

func (k Kind) Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	return k.Wrap(err, err.Error())
}

type skyerr struct {
	msg     string
	kind    Kind
	details Details
}

func (e *skyerr) Error() string { return e.msg }
func (e *skyerr) FillDetails(d errorutil.Details) {
	for key, value := range e.details {
		d[key] = value
	}
}

func AddDetails(err error, d errorutil.Details) error {
	var e *skyerr
	if errors.As(err, &e) {
		for key, value := range d {
			e.details[key] = value
		}
	}

	return err
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

func mergeInfo(infos ...map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, info := range infos {
		for k, v := range info {
			out[k] = v
		}
	}
	return out
}

func AsAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	details := errorutil.CollectDetails(err, nil)
	info := errorutil.FilterDetails(details, APIErrorDetail)

	var apiError *APIError
	if errors.As(err, &apiError) {
		apiError.Info = mergeInfo(apiError.Info, info)
		return apiError
	}

	var e *skyerr
	if errors.As(err, &e) {
		return &APIError{
			Kind:    e.kind,
			Message: e.Error(),
			Code:    e.kind.Name.HTTPStatus(),
			Info:    info,
		}
	}

	var v *validation.AggregatedError
	if errors.As(err, &v) {
		causes := make([]Cause, len(v.Errors))
		for i, c := range v.Errors {
			c := c
			causes[i] = &c
		}
		info["causes"] = causes
		return &APIError{
			Kind:    ValidationFailed,
			Message: v.Message,
			Code:    ValidationFailed.Name.HTTPStatus(),
			Info:    info,
		}
	}

	return &APIError{
		Kind:    Kind{InternalError, "UnexpectedError"},
		Message: "unexpected error occurred",
		Code:    InternalError.HTTPStatus(),
		Info:    info,
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

func NewUnauthorized(msg string) error {
	return Unauthorized.WithReason(string(Unauthorized)).New(msg)
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

func NewTooManyRequest(msg string) error {
	return TooManyRequest.WithReason(string(TooManyRequest)).New(msg)
}
