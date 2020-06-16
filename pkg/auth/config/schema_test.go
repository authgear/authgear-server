package config_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

func TestAppConfigSchema(t *testing.T) {
	testFiles := []string{
		"testdata/messaging_tests.yaml",
		"testdata/hook_tests.yaml",
	}

	type TestCase struct {
		Part  string      `yaml:"part"`
		Name  string      `yaml:"name"`
		Error *string     `yaml:"error"`
		Value interface{} `yaml:"value"`
	}
	var testCases []TestCase
	loadTestCases := func(filename string) {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		decoder := goyaml.NewDecoder(f)
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			testCases = append(testCases, testCase)
		}
	}
	for _, n := range testFiles {
		loadTestCases(n)
	}

	Convey("AppConfig schema parts", t, func() {
		for _, testCase := range testCases {
			name := fmt.Sprintf("%s/%s", testCase.Part, testCase.Name)
			Convey(name, func() {
				data, err := json.Marshal(testCase.Value)
				if err != nil {
					panic(err)
				}

				err = config.Schema.ValidateReaderByPart(bytes.NewReader(data), testCase.Part)
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
