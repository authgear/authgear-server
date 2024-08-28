package main

import (
	"strings"
)

type FinalNewlineRule struct{}

var _ Rule = FinalNewlineRule{}

func (r FinalNewlineRule) Check(content string, path string) LintViolations {
	var violations LintViolations
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

func (r FinalNewlineRule) Key() string {
	return "finalNewline"
}
