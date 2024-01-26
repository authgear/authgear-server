package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type multilineError string

func (e multilineError) Error() string {
	return string(e)
}

type Rule interface {
	Check(content string, path string) []LintViolation
}

type LintViolation struct {
	Path    string
	Line    int
	Column  int
	Message string
}

func (e LintViolation) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.Line, e.Column, e.Message)
}

type Linter struct {
	Path           string
	IgnorePatterns []string
	Rules          []Rule
	Errors         []LintViolation
}

func isGoTemplateFile(info os.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && strings.HasSuffix(name, ".html")
}

func (l *Linter) Lint() (err error) {
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
		violation, err := l.LintFile(path, info)
		if err != nil {
			return err
		}
		if violation != nil {
			l.Errors = append(l.Errors, *violation)
		}
		return nil
	})
	return
}

func (l *Linter) LintFile(path string, info os.FileInfo) (violation *LintViolation, err error) {
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
		l.Errors = append(l.Errors, rule.Check(string(content), path)...)
	}

	return
}

func doMain() (err error) {
	if len(os.Args) < 2 {
		err = fmt.Errorf("usage: gotemplatelinter <path/to/htmls>")
		return
	}
	path := os.Args[1]
	linter := Linter{
		IgnorePatterns: []string{
			"__generated_asset.html",
		},
		Rules: []Rule{
			IndentationRule{},
			FinalNewlineRule{},
		},
		Path: path,
	}
	err = linter.Lint()
	if err != nil {
		return
	}

	if len(linter.Errors) > 0 {
		var buf strings.Builder
		errorsByPath := make(map[string][]error)
		for _, e := range linter.Errors {
			// Assuming e has a Path field
			errorsByPath[e.Path] = append(errorsByPath[e.Path], e)
		}
		for _, errors := range errorsByPath {
			for _, e := range errors {
				fmt.Fprintf(&buf, "%v\n", e)
			}
		}
		fmt.Fprintf(&buf, "\n%d errors found.\n", len(linter.Errors))
		err = multilineError(buf.String())
		return
	}

	return
}

func main() {
	err := doMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("No errors found")
}
