package apierrors

import (
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/util/copyutil"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Details = errorutil.Details

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
	Message string  `json:"message"`
	Code    int     `json:"code"`
	Info    Details `json:"info,omitempty"`
}

var _ error = (*APIError)(nil)
var _ errorutil.Detailer = (*APIError)(nil)
var _ log.LoggingSkippable = (*APIError)(nil)

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *APIError) FillDetails(details Details) {
	for key, val := range e.Info {
		details[key] = val
	}
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

func (e *APIError) SkipLogging() bool {
	return e.Kind.IsSkipLoggingToExternalService
}

func (e *APIError) Clone() *APIError {
	ee, err := copyutil.Clone(e)
	if err != nil {
		return nil
	}
	return ee.(*APIError)
}

// NewWithInfo wraps all value in info with APIErrorDetail, making them appear in the response.
func (k Kind) NewWithInfo(msg string, info Details) error {
	d := Details{}
	for k, v := range info {
		d[k] = APIErrorDetail.Value(v)
	}
	return &APIError{
		Kind:    k,
		Message: msg,
		Code:    k.Name.HTTPStatus(),
		Info:    d,
	}
}

// New is a shorthand of NewWithInfo with an empty Details.
func (k Kind) New(msg string) error {
	return k.NewWithInfo(msg, make(Details))
}

func (k Kind) NewWithCause(msg string, c Cause) error {
	return k.NewWithInfo(msg, Details{"cause": c})
}

func (k Kind) NewWithCauses(msg string, cs []Cause) error {
	return k.NewWithInfo(msg, Details{"causes": cs})
}

func IsAPIError(err error) bool {
	if _, ok := err.(*APIError); ok {
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

	var jsonSyntaxError *json.SyntaxError
	if errors.As(err, &jsonSyntaxError) {
		return newInvalidJSON(jsonSyntaxError)
	}

	var apiError *APIError
	if errors.As(err, &apiError) {
		apiError.Info = mergeInfo(apiError.Info, info)
		return apiError
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
		Kind:    Kind{Name: InternalError, Reason: "UnexpectedError"},
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

func newInvalidJSON(err *json.SyntaxError) *APIError {
	return &APIError{
		Kind:    Kind{Name: BadRequest, Reason: "InvalidJSON"},
		Message: err.Error(),
		Code:    BadRequest.HTTPStatus(),
		Info: map[string]interface{}{
			"byte_offset": err.Offset,
		},
	}
}
