package cmdinternal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrateSetDefaultLogoHeight(t *testing.T) {
	Convey("migrateSetDefaultLogoHeight", t, func() {
		test := func(srcJSON string, expectedOutputJSON string, expectedErr error) {
			src := make(map[string]string)
			err := json.Unmarshal([]byte(srcJSON), &src)
			if err != nil {
				panic(err)
			}
			expectedOutput := make(map[string]string)
			err = json.Unmarshal([]byte(expectedOutputJSON), &expectedOutput)
			if err != nil {
				panic(err)
			}
			err = migrateSetDefaultLogoHeight("dummy-app-id", src, false)
			So(err, ShouldResemble, expectedErr)
			So(src, ShouldResemble, expectedOutput) // src was modified in-place
		}

		toB64 := func(str string) string {
			return base64.StdEncoding.EncodeToString([]byte(str))
		}

		lightThemeCSSWithLogoHeightOnly := `:root {
  --brand-logo__height: 40px;
}
`
		darkThemeCSSWithLogoHeightOnly := `:root.dark {
  --brand-logo__height: 40px;
}
`
		b64LightCSS := toB64(lightThemeCSSWithLogoHeightOnly)
		b64DarkCSS := toB64(darkThemeCSSWithLogoHeightOnly)
		Convey("!hasLightLogo && !hasLightThemeCSS && !hasDarkLogo && !hasDarkThemeCSS", func() {
			Convey("should do nothing", func() {
				test(
					`{}`,
					`{}`,
					nil,
				)
			})
		})

		Convey("hasLightLogo && !hasLightThemeCSS", func() {
			Convey("should create light-theme.css", func() {
				test(
					`{
  "static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img"
}`,
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, b64LightCSS),
					nil,
				)
			})
		})
		Convey("hasLightLogo && hasLightThemeCSS && alreadySet", func() {
			Convey("should do nothing", func() {
				test(
					fmt.Sprintf(`{
  "static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, b64LightCSS),
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, b64LightCSS),
					nil,
				)
			})
		})
		Convey("hasLightLogo && hasLightThemeCSS && notAlreadySet", func() {
			Convey("should add logo height property", func() {
				originalCSS := toB64(`:root {
  --layout__bg-color: #0047AB;
}
`)
				outputCSS := toB64(`:root {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`)
				test(
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, originalCSS),
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, outputCSS),
					nil,
				)
			})
		})

		Convey("!hasLightLogo && hasLightThemeCSS", func() {
			Convey("should do nothing because logo not set", func() {
				originalCSS := toB64(`:root {
  --layout__bg-color: #0047AB;
}
`)
				test(
					fmt.Sprintf(`{
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, originalCSS),
					fmt.Sprintf(`{
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v"
}`, originalCSS),
					nil,
				)
			})
		})

		Convey("hasDarkLogo && !hasDarkThemeCSS", func() {
			Convey("should create dark-theme.css", func() {
				test(
					`{
  "static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img"
}`,
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, b64DarkCSS),
					nil,
				)
			})
		})
		Convey("hasDarkLogo && hasDarkThemeCSS && alreadySet", func() {
			Convey("should do nothing", func() {
				test(
					fmt.Sprintf(`{
  "static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, b64DarkCSS),
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, b64DarkCSS),
					nil,
				)
			})
		})
		Convey("hasDarkLogo && hasDarkThemeCSS && notAlreadySet", func() {
			Convey("should add logo height property", func() {
				originalCSS := toB64(`:root.dark {
  --layout__bg-color: #0047AB;
}
`)
				outputCSS := toB64(`:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`)
				test(
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, originalCSS),
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, outputCSS),
					nil,
				)
			})
		})
		Convey("!hasDarkLogo && hasDarkThemeCSS", func() {
			Convey("should do nothing because logo not set", func() {
				originalCSS := toB64(`:root.dark {
  --layout__bg-color: #0047AB;
}
`)
				test(
					fmt.Sprintf(`{
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, originalCSS),
					fmt.Sprintf(`{
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, originalCSS),
					nil,
				)
			})
		})

		Convey("hasLightLogo && hasLightThemeCSS && notAlreadySet && hasDarkLogo && hasDarkThemeCSS && alreadySet", func() {
			Convey("should modify light css and do nothing to dark css", func() {
				originalLightCSS := toB64(`:root {
  --layout__bg-color: #0047AB;
}
`)
				outputLightCSS := toB64(`:root {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`)
				originalDarkCSS := toB64(`:root.dark {
  --layout__bg-color: #0047AB;
	--brand-logo__height: 40px;
}
`)
				test(
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v",
	"static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, originalLightCSS, originalDarkCSS),
					fmt.Sprintf(`{
	"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-light-theme.css": "%v",
	"static_2f_zh-HK_2f_app_5f_logo_dark.png": "base64-encoded-img",
	"static_2f_authgear-authflowv_32_-dark-theme.css": "%v"
}`, outputLightCSS, originalDarkCSS),
					nil,
				)
			})

		})

	})

	Convey("getMatchingConfigSourcePaths", t, func() {
		test := func(regex *regexp.Regexp, configSourceData map[string]string, expectedMatched []string, expectedOK bool) {
			matched, ok := getMatchingConfigSourcePaths(regex, configSourceData)

			sortedMatched := sort.StringSlice(matched)
			sortedMatched.Sort()

			sortedExpectedMatched := sort.StringSlice(expectedMatched)
			sortedExpectedMatched.Sort()
			So(sortedMatched, ShouldResemble, sortedExpectedMatched)
			So(ok, ShouldEqual, expectedOK)
		}

		configSourceFixture := map[string]string{
			// light logos
			"static_2f_zh-HK_2f_app_5f_logo.png": "base64-encoded",
			"static_2f_en_2f_app_5f_logo.png":    "base64-encoded",
			"static_2f_es-ES_2f_app_5f_logo.png": "base64-encoded",

			// dark logos
			"static_2f_zh-HK_2f_app_5f_logo_5f_dark.png": "base64-encoded",
			"static_2f_en_2f_app_5f_logo_5f_dark.png":    "base64-encoded",
			"static_2f_es-ES_2f_app_5f_logo_5f_dark.png": "base64-encoded",

			// light theme css
			"static_2f_authgear-authflowv_32_-light-theme.css": "base64-encoded",

			// dark theme css
			"static_2f_authgear-authflowv_32_-dark-theme.css": "base64-encoded",

			// true-negatives
			"authgear.yaml":                          "base64-encoded",
			"authgear.secrets.yaml":                  "base64-encoded",
			"templates_2f_zh-HK_2f_translation.json": "base64-encoded",
		}

		Convey("nothing matched", func() {
			var expectedMatched []string
			expectedOK := false

			test(regexp.MustCompile(`^wrong regex$`), configSourceFixture, expectedMatched, expectedOK)
		})

		Convey("matched 3 light logos", func() {
			expectedMatched := []string{
				"static_2f_en_2f_app_5f_logo.png",
				"static_2f_es-ES_2f_app_5f_logo.png",
				"static_2f_zh-HK_2f_app_5f_logo.png",
			}
			expectedOK := true
			test(regexp.MustCompile(`^static/([a-zA-Z-]+)/app_logo\.(png|jpe|jpeg|jpg|gif)$`), configSourceFixture, expectedMatched, expectedOK)
		})

		Convey("matched 3 dark logos", func() {
			expectedMatched := []string{
				"static_2f_zh-HK_2f_app_5f_logo_5f_dark.png",
				"static_2f_en_2f_app_5f_logo_5f_dark.png",
				"static_2f_es-ES_2f_app_5f_logo_5f_dark.png",
			}
			expectedOK := true
			test(regexp.MustCompile(`^static/([a-zA-Z-]+)/app_logo_dark\.(png|jpe|jpeg|jpg|gif)$`), configSourceFixture, expectedMatched, expectedOK)
		})

		Convey("matched 1 light theme css", func() {
			expectedMatched := []string{
				"static_2f_authgear-authflowv_32_-light-theme.css",
			}
			expectedOK := true
			test(regexp.MustCompile(`^static/authgear-authflowv2-light-theme.css$`), configSourceFixture, expectedMatched, expectedOK)
		})

		Convey("matched 1 dark theme css", func() {
			expectedMatched := []string{
				"static_2f_authgear-authflowv_32_-dark-theme.css",
			}
			expectedOK := true
			test(regexp.MustCompile(`^static/authgear-authflowv2-dark-theme.css$`), configSourceFixture, expectedMatched, expectedOK)
		})

	})
}
