package validation

import (
	"fmt"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var defaultErrorMessage = "invalid value"

var ValidationFailed = skyerr.Invalid.WithReason("ValidationFailed")

type Error struct {
	Location string                 `json:"location"`
	Keyword  string                 `json:"kind"`
	Info     map[string]interface{} `json:"details,omitempty"`
}

func (e *Error) Kind() string { return e.Keyword }

func (e *Error) String() string {
	loc := e.Location
	if loc == "" {
		loc = "<root>"
	}

	if e.Keyword == "general" {
		msg, _ := e.Info["msg"].(string)
		return fmt.Sprintf("%s: %s", loc, msg)
	}
	if e.Info == nil {
		return fmt.Sprintf("%s: %s", loc, e.Keyword)
	}
	return fmt.Sprintf("%s: %s\n  %v", loc, e.Keyword, e.Info)
}

type AggregatedError struct {
	Message string
	Errors  []Error
}

func (e *AggregatedError) Error() string {
	lines := []string{e.Message + ":"}
	for _, err := range e.Errors {
		lines = append(lines, err.String())
	}
	return strings.Join(lines, "\n")
}

func (e *AggregatedError) AsAPIError() *skyerr.APIError {
	return &skyerr.APIError{
		Kind:    ValidationFailed,
		Message: e.Message,
		Code:    ValidationFailed.Name.HTTPStatus(),
		Info: map[string]interface{}{
			"causes": e.Errors,
		},
	}
}
