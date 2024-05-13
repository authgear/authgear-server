package config_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestParseFeatureConfig(t *testing.T) {
	Convey("default feature config", t, func() {
		cfg, err := config.ParseFeatureConfig([]byte(`{}`))
		So(err, ShouldBeNil)

		data, err := ioutil.ReadFile("testdata/default_feature.yaml")
		So(err, ShouldBeNil)

		var defaultCfg config.FeatureConfig
		err = yaml.Unmarshal(data, &defaultCfg)
		So(err, ShouldBeNil)

		So(cfg, ShouldResemble, &defaultCfg)
	})

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
