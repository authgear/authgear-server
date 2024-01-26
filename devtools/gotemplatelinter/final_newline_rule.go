package main

import (
	"strings"
)

type FinalNewlineRule struct{}

func (r FinalNewlineRule) Check(content string, path string) []LintViolation {
	var violations []LintViolation
	lines := strings.Split(content, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lineNumber := len(lines)
		columnNumber := len(lines[lineNumber-1]) + 1
		violations = append(violations, LintViolation{
			Path:    path,
			Line:    lineNumber,
			Column:  columnNumber,
			Message: "File does not end with a newline (final-newline)",
		})
	}
	return violations
}
