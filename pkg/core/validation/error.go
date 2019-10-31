package validation

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ValidationFailed = skyerr.Invalid.WithReason("ValidationFailed")

type skyerrCause struct {
	ErrorCause
}

func (c skyerrCause) Kind() string { return string(c.ErrorCause.Kind) }

func NewValidationFailed(msg string, causes []ErrorCause) error {
	ecauses := make([]skyerr.Cause, len(causes))
	for i, c := range causes {
		ecauses[i] = skyerrCause{c}
	}
	return ValidationFailed.NewWithCauses(msg, ecauses)
}

func ErrorCauses(err error) []ErrorCause {
	apiError := skyerr.AsAPIError(err)
	causes, _ := apiError.Info["causes"].([]skyerr.Cause)

	var s []ErrorCause
	for _, c := range causes {
		if c, ok := c.(skyerrCause); ok {
			s = append(s, c.ErrorCause)
		}
	}
	return s
}

func ErrorCauseStrings(err error) []string {
	causes := ErrorCauses(err)
	s := make([]string, len(causes))
	for i, c := range causes {
		s[i] = c.String()
	}
	return s
}
