package main

import (
	"strings"
	"unicode"
)

type IndentationRule struct{}

var _ Rule = IndentationRule{}

func (r IndentationRule) Check(content string, path string) LintViolations {
	var violations LintViolations
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		indent := len(line) - len(strings.TrimFunc(line, unicode.IsSpace))
		firstTab := strings.IndexRune(line[:indent], '\t')
		if firstTab != -1 {
			violations = append(violations, LintViolation{
				Path:    path,
				Line:    i + 1,
				Column:  firstTab + 1,
				Message: "Indentation is a tab instead of 2 spaces (indent)",
			})
		}
	}
	return violations
}

func (r IndentationRule) Key() string {
	return "indentation"
}
