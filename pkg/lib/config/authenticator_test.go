package config_test

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAuthenticatorCodeValidPeriodDeprecation(t *testing.T) {
  Convey("authenticator_code_valid_period_deprecation", t, func() {
    entries, err := os.ReadDir("testdata/authenticator_code_valid_period_deprecation/input")
    if err != nil {
      panic(err)
    }
    for _, entry := range entries {
      withoutYAML := entry.Name()[:len(entry.Name())-5]
      withoutInput := withoutYAML[:len(withoutYAML)-6]
      Convey(withoutInput, func() {
        inputFilePath := "testdata/authenticator_code_valid_period_deprecation/input/" + entry.Name()
        input, err := os.ReadFile(inputFilePath)
        if err != nil {
          panic(err)
        }

        cfg, err := config.Parse(input)
        So(err, ShouldBeNil)

        outputEntry := withoutInput + ".output.yaml"
        outputFilePath := "testdata/authenticator_code_valid_period_deprecation/output/" + outputEntry
        data, err := os.ReadFile(outputFilePath)
        if err != nil {
          panic(err)
        }

        expected, err := config.Parse(data)
        So(err, ShouldBeNil)

        So(cfg, ShouldResemble, expected)
      })
    }
  })
}