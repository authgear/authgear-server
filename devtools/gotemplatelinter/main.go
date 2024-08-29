package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/slice"
)

type Rule interface {
	Check(content string, path string) LintViolations
	Key() string
}

type LintViolation struct {
	Path    string
	Line    int
	Column  int
	Message string
}

type LintViolations []LintViolation

func (violations LintViolations) Error() string {
	var buf strings.Builder
	violationsByPath := make(map[string]LintViolations)
	for _, v := range violations {
		violationsByPath[v.Path] = append(violationsByPath[v.Path], v)
	}

	for _, violations := range violationsByPath {
		for _, v := range violations {
			fmt.Fprintf(&buf, "%s:%d:%d: %s\n", v.Path, v.Line, v.Column, v.Message)
		}
	}

	return buf.String()
}

type Linter struct {
	Path           string
	IgnorePatterns []string
	Rules          []Rule
}

func isGoTemplateFile(info os.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && strings.HasSuffix(name, ".html")
}

func (l *Linter) Lint() (violations LintViolations, err error) {
	err = filepath.Walk(l.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isGoTemplateFile(info) {
			return nil
		}
		for _, pattern := range l.IgnorePatterns {
			if filepath.Base(path) == pattern {
				return nil
			}

			matched, err := filepath.Match(pattern, path)
			if err != nil {
				return err
			}
			if matched {
				return nil
			}
		}
		fileViolations, err := l.LintFile(path, info)
		if err != nil {
			return err
		}
		violations = append(violations, fileViolations...)
		return nil
	})
	return
}

func (l *Linter) LintFile(path string, info os.FileInfo) (violations LintViolations, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return
	}

	for _, rule := range l.Rules {
		violations = append(violations, rule.Check(string(content), path)...)
	}

	return
}

func constructRules(rulesToIgnore []string) []Rule {
	indentationRule := IndentationRule{}
	EOLAtEOFRule := EOLAtEOFRule{}
	translationKeyRule := TranslationKeyRule{}
	rules := []Rule{
		indentationRule,
		EOLAtEOFRule,
		translationKeyRule,
	}
	ignoreRuleFn := func(rule Rule) {
		rules = slice.Filter[Rule](rules, func(r Rule) bool {
			return r != rule
		})
	}
	for _, ruleToIgnore := range rulesToIgnore {
		switch ruleToIgnore {
		case indentationRule.Key():
			ignoreRuleFn(indentationRule)
		case EOLAtEOFRule.Key():
			ignoreRuleFn(EOLAtEOFRule)
		case translationKeyRule.Key():
			ignoreRuleFn(translationKeyRule)
		}
	}

	return rules
}

func doMain() (violations LintViolations, err error) {
	if len(os.Args) < 2 {
		err = fmt.Errorf("usage: gotemplatelinter --path <path/to/htmls> --ignore-rule rule1ToIgnore --ignore-rule rule2ToIgnore")
		return
	}
	flags := ParseFlags()
	rules := constructRules(flags.RulesToIgnore)
	linter := Linter{
		IgnorePatterns: []string{
			"__generated_asset.html",
		},
		Rules: rules,
		Path:  flags.Path,
	}
	violations, err = linter.Lint()
	if err != nil {
		return
	}

	return
}

func main() {
	violations, err := doMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if len(violations) > 0 {
		var err error
		err = violations
		fmt.Fprintf(os.Stderr, "%v", err)
		fmt.Fprintf(os.Stderr, "%v errors found\n", len(violations))
		os.Exit(1)
	}
}
