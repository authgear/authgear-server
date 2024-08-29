package main

import (
	"strings"
)

type EOLAtEOFRule struct{}

var _ Rule = EOLAtEOFRule{}

func (r EOLAtEOFRule) Check(content string, path string) LintViolations {
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

func (r EOLAtEOFRule) Key() string {
	return "eol-at-eof"
}
