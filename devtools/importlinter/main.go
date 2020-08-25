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

func getPackageName(path string) string {
	return filepath.Dir(filepath.Join("github.com/authgear/authgear-server", path))
}

func isModulePackage(packageName string) bool {
	return strings.HasPrefix(packageName, "github.com/authgear/authgear-server/pkg")
}

func getPackageCategory(packageName string) (cat string, err error) {
	rel, err := filepath.Rel("github.com/authgear/authgear-server/pkg", packageName)
	if err != nil {
		return
	}

	parts := strings.Split(rel, "/")
	cat = parts[0]
	return
}

type Linter struct {
	PackageCategory        string
	AllowedPackageCategory []string
	Errors                 []error
}

func (l *Linter) Lint() (err error) {
	err = filepath.Walk("./pkg", func(path string, info os.FileInfo, err error) error {
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

	packageName := getPackageName(path)
	packageCategory, err := getPackageCategory(packageName)
	if err != nil {
		return
	}

	// This is not our target. Skip it.
	if packageCategory != l.PackageCategory {
		return
	}

	allowed := make(map[string]struct{})
	// Always allow importing self.
	allowed[l.PackageCategory] = struct{}{}
	for _, a := range l.AllowedPackageCategory {
		allowed[a] = struct{}{}
	}

	for _, s := range astFile.Imports {
		// Remove the quotes around the import path.
		importPath := s.Path.Value[1 : len(s.Path.Value)-1]
		if !isModulePackage(importPath) {
			continue
		}

		packageCategory, err = getPackageCategory(importPath)
		if err != nil {
			return
		}
		_, ok := allowed[packageCategory]
		if !ok {
			violation = fmt.Errorf("%s cannot import %s", path, importPath)
		}
	}

	return
}

func doMain() (err error) {
	if len(os.Args) < 3 {
		err = fmt.Errorf("usage: lintimports <pkg> [allowed]...")
		return
	}
	pkg := os.Args[1]
	allowed := os.Args[2:]
	linter := Linter{
		PackageCategory:        pkg,
		AllowedPackageCategory: allowed,
	}
	err = linter.Lint()
	if err != nil {
		return
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
