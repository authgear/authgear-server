package config_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAuthenticatorCodeValidPeriodDeprecation(t *testing.T) {
	Convey("authenticator_code_valid_period_deprecation", t, func() {
		// filter out files that are input files with Glob function
		filteredEntries, err := filepath.Glob("testdata/authenticator_code_valid_period_deprecation/*.input.yaml")
		if err != nil {
			panic(err)
		}

		for _, entry := range filteredEntries {
			withoutYAML := entry[:len(entry)-5]
			withoutInput := withoutYAML[:len(withoutYAML)-6]
			Convey(withoutInput, func() {
				input, err := os.ReadFile(entry)
				if err != nil {
					panic(err)
				}

				cfg, err := config.Parse(input)
				So(err, ShouldBeNil)

				outputFilePath := withoutInput + ".output.yaml"
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
