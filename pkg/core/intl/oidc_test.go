package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalizeJSONObject(t *testing.T) {
	Convey("LocalizeJSONObject", t, func() {
		Convey("simple", func() {
			jsonObject := map[string]interface{}{
				"client_name":            "client_name default",
				"client_name#zh":         "client_name zh",
				"client_name#zh-Hant-HK": "client_name zh-Hant-HK",
			}

			test := func(tags []string, expected string) {
				value := LocalizeJSONObject(tags, "", jsonObject, "client_name")
				So(value, ShouldEqual, expected)
			}

			test(nil, "client_name default")
			test([]string{"en"}, "client_name default")
			test([]string{"zh-Hant-HK"}, "client_name zh-Hant-HK")
		})

		Convey("invalid fallback", func() {
			jsonObject := map[string]interface{}{
				"client_name#zh":         "client_name zh",
				"client_name#zh-Hant-HK": "client_name zh-Hant-HK",
			}

			test := func(tags []string, expected string) {
				value := LocalizeJSONObject(tags, "", jsonObject, "client_name")
				So(value, ShouldEqual, expected)
			}

			test(nil, "")
			test([]string{"en"}, "")
			test([]string{"zh-Hant-HK"}, "client_name zh-Hant-HK")
		})
	})
}

func TestLocalizeStringMap(t *testing.T) {
	Convey("LocalizeStringMap", t, func() {
		Convey("simple case", func() {
			stringMap := map[string]string{
				"subject":            "subject default",
				"subject#zh":         "subject zh",
				"subject#zh-Hant-HK": "subject zh-Hant-HK",
			}

			test := func(tags []string, expected string) {
				value := LocalizeStringMap(tags, "", stringMap, "subject")
				So(value, ShouldEqual, expected)
			}

			test(nil, "subject default")
			test([]string{"en"}, "subject default")
			test([]string{"zh-Hant-HK"}, "subject zh-Hant-HK")
		})

		Convey("invalid fallback", func() {
			stringMap := map[string]string{
				"subject#zh":         "subject zh",
				"subject#zh-Hant-HK": "subject zh-Hant-HK",
			}

			test := func(tags []string, expected string) {
				value := LocalizeStringMap(tags, "", stringMap, "subject")
				So(value, ShouldEqual, expected)
			}

			test(nil, "")
			test([]string{"en"}, "")
			test([]string{"zh-Hant-HK"}, "subject zh-Hant-HK")
		})
	})
}
