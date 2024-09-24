package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
)

func TestMessagingSMSGateway(t *testing.T) {
	schema := Schema.PartValidator("SMSGatewayConfig")

	Convey("ParseMessagingSMSGateway", t, func() {
		f, err := os.Open("testdata/parse_messaging_sms_gateway_tests.yaml")
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
				jsonData, err := yaml.YAMLToJSON(data)
				if err != nil {
					panic(err)
				}

				err = schema.Validate(bytes.NewReader(jsonData))
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
