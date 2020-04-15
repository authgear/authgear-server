package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalize(t *testing.T) {
	Convey("Localize", t, func() {
		test := func(m map[string]string, tags []string, expected string) {
			_, value := Localize(tags, m)
			So(value, ShouldEqual, expected)
		}

		// Select default if there is no preferred languages
		test(map[string]string{
			"":   "Hello from default",
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, nil, "Hello from default")

		// Select default if there is no preferred languages
		test(map[string]string{
			"":   "Hello from default",
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, []string{}, "Hello from default")

		// Simply select japanese
		test(map[string]string{
			"":   "Hello from default",
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, []string{"ja-JP", "en-US", "zh-Hant-HK"}, "Hello from ja")

		// Select the default because korean is not supported
		test(map[string]string{
			"":   "Hello from default",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, []string{"kr-KR"}, "Hello from default")
	})
}

func TestLocalizeJSONObject(t *testing.T) {
	Convey("LocalizeJSONObject", t, func() {
		jsonObject := map[string]interface{}{
			"client_name":            "client_name default",
			"client_name#en":         "client_name en",
			"client_name#zh":         "client_name zh",
			"client_name#zh-Hant-HK": "client_name zh-Hant-HK",
		}

		test := func(tags []string, expected string) {
			value := LocalizeJSONObject(tags, jsonObject, "client_name")
			So(value, ShouldEqual, expected)
		}

		test(nil, "client_name default")
		test([]string{"en"}, "client_name en")
		test([]string{"zh-Hant-HK"}, "client_name zh-Hant-HK")
	})
}

func TestLocalizeStringMap(t *testing.T) {
	Convey("LocalizeStringMap", t, func() {
		stringMap := map[string]string{
			"subject":            "subject default",
			"subject#en":         "subject en",
			"subject#zh":         "subject zh",
			"subject#zh-Hant-HK": "subject zh-Hant-HK",
		}

		test := func(tags []string, expected string) {
			value := LocalizeStringMap(tags, stringMap, "subject")
			So(value, ShouldEqual, expected)
		}

		test(nil, "subject default")
		test([]string{"en"}, "subject en")
		test([]string{"zh-Hant-HK"}, "subject zh-Hant-HK")
	})
}
