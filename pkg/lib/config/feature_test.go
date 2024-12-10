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

	Convey("merge feature config", t, func() {

		type TestCase struct {
			Configs []interface{} `yaml:"configs"`
			Result  interface{}   `yaml:"result"`
		}

		f, err := os.Open("testdata/merge_feature.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		decoder := goyaml.NewDecoder(f)
		var testCase TestCase
		err = decoder.Decode(&testCase)
		if err != nil {
			panic(err)
		}

		resultData, err := goyaml.Marshal(testCase.Result)
		if err != nil {
			panic(err)
		}

		expected, err := config.ParseFeatureConfig(resultData)
		So(err, ShouldBeNil)

		mergedConfig := &config.FeatureConfig{}
		for _, cfgData := range testCase.Configs {
			data, err := goyaml.Marshal(cfgData)
			if err != nil {
				panic(err)
			}

			cfg, err := config.ParseFeatureConfigWithoutDefaults(data)
			So(err, ShouldBeNil)

			mergedConfig = mergedConfig.Merge(cfg)
		}
		config.SetFieldDefaults(mergedConfig)

		So(mergedConfig, ShouldResemble, expected)
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
