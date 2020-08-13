package validation

import (
	"fmt"
	"strings"
)

var defaultErrorMessage = "invalid value"

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
