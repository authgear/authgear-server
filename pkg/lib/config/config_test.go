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
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func TestAppConfig(t *testing.T) {
	Convey("AppConfig", t, func() {
		minimalAppConfig := `{ "id": "test", "http": { "public_origin": "http://test" } }`

		Convey("populate default values", func() {
			cfg, err := config.Parse([]byte(minimalAppConfig))
			So(err, ShouldBeNil)

			data, err := ioutil.ReadFile("testdata/default_config.yaml")
			if err != nil {
				panic(err)
			}

			var defaultCfg config.AppConfig
			err = yaml.Unmarshal(data, &defaultCfg)
			if err != nil {
				panic(err)
			}

			So(cfg, ShouldResemble, &defaultCfg)
		})

		Convey("round-trip default configuration", func() {
			cfg, err := config.Parse([]byte(minimalAppConfig))
			So(err, ShouldBeNil)

			data, err := yaml.Marshal(cfg)
			So(err, ShouldBeNil)

			cfg2, err := config.Parse(data)
			So(err, ShouldBeNil)
			So(cfg, ShouldResemble, cfg2)
		})

		Convey("parse validation", func() {
			f, err := os.Open("testdata/config_tests.yaml")
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

					_, err = config.Parse(data)
					if testCase.Error != nil {
						So(err, ShouldBeError, *testCase.Error)
					} else {
						So(err, ShouldBeNil)
					}
				})
			}
		})

		Convey("get phone input country", func() {
			data, err := ioutil.ReadFile("testdata/phone_input_config_test.yaml")
			if err != nil {
				panic(err)
			}

			var cfg config.AppConfig
			err = yaml.Unmarshal(data, &cfg)
			if err != nil {
				panic(err)
			}

			countries := cfg.UI.PhoneInput.GetCountries()
			expected := []phone.Country{
				phone.Country{Alpha2: "HK", CountryCallingCode: "852"},
				phone.Country{Alpha2: "US", CountryCallingCode: "1"},
				phone.Country{Alpha2: "GB", CountryCallingCode: "44"},
				phone.Country{Alpha2: "TW", CountryCallingCode: "886"},
			}
			So(countries, ShouldResemble, expected)
		})
	})
}
