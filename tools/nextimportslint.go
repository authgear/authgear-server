package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	gearBasePath                 = "github.com/skygeario/skygear-server/pkg/"
	tolerantImportsCommentPrefix = "// tolerant nextimportslint:"
)

var (
	exitCode = 0
)

func report(err error) {
	scanner.PrintError(os.Stderr, err)
	exitCode = 2
}

func isGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func isTestFile(f os.FileInfo) bool {
	// ignore test files
	name := f.Name()
	return strings.HasSuffix(name, "_test.go")
}

func getTolerantImports(gearName string, comments []*ast.CommentGroup) []string {
	// core is default allow imports
	rules := []string{
		"core",
		// use skyerr a lot, tolerance import server temporarily
		"server",
	}

	// add except rules, like: "// tolerant nextimportslint: record, chat"
	for _, commentGroup := range comments {
		for _, comment := range commentGroup.List {
			if strings.HasPrefix(comment.Text, tolerantImportsCommentPrefix) {
				tolerantGearsText := strings.Replace(comment.Text, tolerantImportsCommentPrefix, "", 1)
				tolerantGears := strings.Split(tolerantGearsText, ",")
				for _, e := range tolerantGears {
					r := strings.TrimSpace(e)
					if r != "" {
						rules = append(rules, r)
					}
				}
			}
		}
	}

	return rules
}

func isTolerantGear(gearName string, tolerantGears []string) bool {
	for _, t := range tolerantGears {
		if t == gearName {
			return true
		}
	}
	return false
}

func illegalImport(gearName string, importGearName string, tolerantGears []string) bool {
	if gearName != importGearName && !isTolerantGear(importGearName, tolerantGears) {
		return true
	}

	return false
}

func processFile(gearName string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	in := f

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	fset := token.NewFileSet() // positions are relative to fset
	mode := parser.ImportsOnly + parser.ParseComments
	astFile, err := parser.ParseFile(fset, filename, src, mode)
	if err != nil {
		return err
	}
	tolerantImports := getTolerantImports(gearName, astFile.Comments)

	for _, s := range astFile.Imports {
		// s.Path.Value contains \" at head and tail
		cleanImportPath := strings.Replace(s.Path.Value, "\"", "", 2)
		if strings.HasPrefix(cleanImportPath, gearBasePath) {
			importGearRelativePath := strings.Replace(cleanImportPath, gearBasePath, "", 1)
			importGearName := strings.Split(importGearRelativePath, "/")[0]
			if illegalImport(gearName, importGearName, tolerantImports) {
				errMsg := fmt.Sprintf("nextimports: doesn't allow \"%s\" imports another gear \"%s\"", filename, importGearName)
				return errors.New(errMsg)
			}
		}
	}

	return nil
}

func nextImportsMain() {
	err := filepath.Walk("pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// only check files under pkg
		if isGoFile(info) && !isTestFile(info) {
			gearName := strings.Split(path, "/")[1]
			if gearName != "server" {
				// only check skygear next gear files
				// "server" is for skygear v1
				err = processFile(gearName, path)
			}
		}
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		report(err)
		return
	}
}

func main() {
	nextImportsMain()
	os.Exit(exitCode)
}
