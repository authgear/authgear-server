package config_test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"
)

func TestParseFeatureConfig(t *testing.T) {
	Convey("ParseFeatureConfig", t, func() {
		f, err := os.Open("testdata/parse_feature_tests.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		type TestCase struct {
			Name   string      `yaml:"name"`
			Error  *string     `yaml:"error"`
			Config interface{} `yaml:"config"`
		}

		decoder := goyaml.NewDecoder(f)
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			Convey(testCase.Name, func() {
				data, err := goyaml.Marshal(testCase.Config)
				if err != nil {
					panic(err)
				}

				_, err = config.ParseFeatureConfig(data)
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
