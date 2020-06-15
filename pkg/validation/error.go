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
	if e.Info == nil {
		return fmt.Sprintf("%s: %s", e.Location, e.Keyword)
	}
	return fmt.Sprintf("%s: %s\n  %v", e.Location, e.Keyword, e.Info)
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
