package validation

import (
	"fmt"
	"strings"
)

type Error struct {
	Location string
	Keyword  string
	Info     map[string]interface{}
}

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
	Errors []Error
}

func (e *AggregatedError) Error() string {
	lines := []string{"invalid JSON:"}
	for _, err := range e.Errors {
		lines = append(lines, err.String())
	}
	return strings.Join(lines, "\n")
}
