package config_test

import (
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"

	"github.com/authgear/authgear-server/pkg/lib/config"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
)

func TestParseSecret(t *testing.T) {
	Convey("ParseSecret", t, func() {
		f, err := os.Open("testdata/parse_secret_tests.yaml")
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

				_, err = config.ParseSecret(data)
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}

func TestSecretConfigValidate(t *testing.T) {
	Convey("SecretConfigValidate", t, func() {
		f, err := os.Open("testdata/secret_config_validate_tests.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		type TestCase struct {
			Name         string      `yaml:"name"`
			Error        *string     `yaml:"error"`
			AppConfig    interface{} `yaml:"app_config"`
			SecretConfig interface{} `yaml:"secret_config"`
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
				appConfigBytes, err := goyaml.Marshal(testCase.AppConfig)
				if err != nil {
					panic(err)
				}

				secretConfigBytes, err := goyaml.Marshal(testCase.SecretConfig)
				if err != nil {
					panic(err)
				}

				appConfig, err := config.Parse(appConfigBytes)
				if err != nil {
					panic(err)
				}

				secretConfig, err := config.ParseSecret(secretConfigBytes)
				if err != nil {
					panic(err)
				}

				err = secretConfig.Validate(appConfig)
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
