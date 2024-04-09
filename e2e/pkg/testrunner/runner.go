package testrunner

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

type TestRunner struct {
	T    *testing.T
	Path string
}

func NewTestRunner(t *testing.T, path string) *TestRunner {
	return &TestRunner{
		T:    t,
		Path: path,
	}
}

func (tr *TestRunner) Run() {
	var t = tr.T

	testCases, err := tr.loadFromPath(tr.Path)
	if err != nil {
		t.Fatalf("failed to load test cases: %v", err)
	}

	hasFocus := false
	for _, testCase := range testCases {
		if testCase.Focus {
			if hasFocus {
				t.Fatal("multiple focus test cases")
			}

			hasFocus = true
			break
		}
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.FullName(), func(t *testing.T) {
			t.Parallel()

			if hasFocus && !tc.Focus {
				t.Skip("skipping non-focus test case")
				return
			}

			tc.Run(t)
		})
	}
}

func (tr *TestRunner) loadFromPath(path string) ([]TestCase, error) {
	var testCases []TestCase
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		decoder := yaml.NewDecoder(strings.NewReader(string(data)))
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			testCase.Path = path
			testCases = append(testCases, testCase)
		}

		return nil
	})

	return testCases, err
}
