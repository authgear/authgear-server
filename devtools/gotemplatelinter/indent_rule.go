package main

import (
	"strings"
)

type IndentationRule struct{}

func (r IndentationRule) Check(content string, path string) []LintViolation {
	var violations []LintViolation
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "\t") {
			violations = append(violations, LintViolation{
				Path:    path,
				Line:    i + 1,
				Column:  1,
				Message: "Indentation is a tab instead of 2 spaces (indent)",
			})
		}
	}
	return violations
}
