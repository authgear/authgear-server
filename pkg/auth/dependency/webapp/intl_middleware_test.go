package webapp

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPreferredLanguageTagsFromRequest(t *testing.T) {
	Convey("PreferredLanguageTagsFromRequest", t, func() {
		test := func(uiLocales string, acceptLanguage string, expected []string) {
			r, _ := http.NewRequest("GET", "http://example.com", nil)
			q := url.Values{}
			if uiLocales != "" {
				q.Set("ui_locales", uiLocales)
				r.URL.RawQuery = q.Encode()
			}
			if acceptLanguage != "" {
				r.Header.Set("Accept-Language", acceptLanguage)
			}
			actual := PreferredLanguageTagsFromRequest(r)
			So(actual, ShouldResemble, expected)
		}

		// No ui_locales or Accept-Language
		test("", "", nil)

		// Accept-Language
		test("", "zh-Hant-HK; q=0.5, en", []string{"en", "zh-Hant-HK"})

		// ui_locales
		test("ja-JP zh-Hant-TW", "zh-Hant-HK; q=0.5, en", []string{"ja-JP", "zh-Hant-TW"})
	})
}
