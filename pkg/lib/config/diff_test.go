package config_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestDiffAppConfig(t *testing.T) {
	Convey("DiffAppConfig", t, func() {

		Convey("generate diff", func() {
			baseCfg := config.AppConfig{
				ID: "test",
				Localization: &config.LocalizationConfig{
					SupportedLanguages: []string{"en-US"},
				},
			}
			newCfg := config.AppConfig{
				ID: "test",
				Localization: &config.LocalizationConfig{
					SupportedLanguages: []string{"zh-HK-Hant"},
				},
				UI: &config.UIConfig{
					DarkThemeDisabled: true,
				},
			}
			diff, err := config.DiffAppConfig(&baseCfg, &newCfg)
			So(err, ShouldBeNil)

			fmt.Println(diff)

			data, err := ioutil.ReadFile("testdata/diff.txt")
			if err != nil {
				panic(err)
			}

			So(diff, ShouldEqual, string(data))
		})

	})
}
