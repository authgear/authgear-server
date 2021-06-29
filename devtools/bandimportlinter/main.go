package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type multilineError string

func (e multilineError) Error() string {
	return string(e)
}

func isGoFile(info os.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && strings.HasSuffix(name, ".go")
}

type Linter struct {
	BandPackages map[string]struct{}
	Errors       []error
}

func NewLinter() (*Linter, error) {
	contentbytes, err := ioutil.ReadFile(".bandimportpackages")
	if err != nil {
		return nil, fmt.Errorf("failed to load .bandimportpackages, %w", err)
	}

	content := string(contentbytes)
	packages := strings.Split(content, "\n")

	linter := &Linter{
		BandPackages: map[string]struct{}{},
	}
	for _, p := range packages {
		if pkg := strings.TrimSpace(p); pkg != "" {
			linter.BandPackages[pkg] = struct{}{}
		}
	}
	return linter, nil
}

func (l *Linter) Lint(pkgFolder string) (err error) {
	err = filepath.Walk(pkgFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isGoFile(info) {
			return nil
		}
		violation, err := l.LintFile(path, info)
		if err != nil {
			return err
		}
		if violation != nil {
			l.Errors = append(l.Errors, violation)
		}
		return nil
	})
	return
}

func (l *Linter) LintFile(path string, info os.FileInfo) (violation error, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	fileSet := token.NewFileSet()
	mode := parser.ImportsOnly + parser.ParseComments
	astFile, err := parser.ParseFile(fileSet, path, content, mode)
	if err != nil {
		return
	}

	for _, s := range astFile.Imports {
		importPath := s.Path.Value[1 : len(s.Path.Value)-1]
		if _, ok := l.BandPackages[importPath]; ok {
			violation = fmt.Errorf("%s cannot import %s", path, importPath)
		}
	}
	return
}

func doMain() (err error) {
	if len(os.Args) < 1 {
		err = fmt.Errorf("usage: bandimportlinter [packages]...")
		return
	}
	linter, err := NewLinter()
	if err != nil {
		return
	}
	for _, pkgFolder := range os.Args {
		err = linter.Lint(pkgFolder)
		if err != nil {
			return
		}
	}

	if len(linter.Errors) > 0 {
		var buf strings.Builder
		for _, e := range linter.Errors {
			fmt.Fprintf(&buf, "%v\n", e)
		}
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
}
