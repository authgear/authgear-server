package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResolveRecaptchaV2(t *testing.T) {
	Convey("ResolveRecaptchaV2", t, func() {
		test := func(target string, expected string) {
			actual := ResolveRecaptchaV2(target)
			So(actual, ShouldEqual, expected)
			So(RecaptchaV2Languages, ShouldContain, actual)
		}

		// authgear-supported, non-recaptcha-supported
		test("es-ES", "es")

		// both authgear- & recaptcha-supported
		test("en", "en")
		test("en", "en")
		test("af", "af")
		test("am", "am")
		test("ar", "ar")
		test("hy", "hy")
		test("az", "az")
		test("eu", "eu")
		test("be", "be")
		test("bn", "bn")
		test("bg", "bg")
		test("my", "my")
		test("ca", "ca")
		test("zh-HK", "zh-HK")
		test("zh-CN", "zh-CN")
		test("zh-TW", "zh-TW")
		test("hr", "hr")
		test("cs", "cs")
		test("da", "da")
		test("nl", "nl")
		test("et", "et")
		test("fil", "fil")
		test("fi", "fi")
		test("fr", "fr")
		test("gl", "gl")
		test("ka", "ka")
		test("de", "de")
		test("el", "el")
		test("hi", "hi")
		test("hu", "hu")
		test("is", "is")
		test("id", "id")
		test("it", "it")
		test("ja", "ja")
		test("kn", "kn")
		test("km", "km")
		test("ko", "ko")
		test("ky", "ky")
		test("lo", "lo")
		test("lv", "lv")
		test("lt", "lt")
		test("mk", "mk")
		test("ms", "ms")
		test("ml", "ml")
		test("mr", "mr")
		test("mn", "mn")
		test("ne", "ne")
		test("no", "no")
		test("fa", "fa")
		test("pl", "pl")
		test("pt-PT", "pt-PT")
		test("pt-BR", "pt-BR")
		test("pt", "pt")
		test("ro", "ro")
		test("ru", "ru")
		test("sr", "sr")
		test("si", "si")
		test("sk", "sk")
		test("sl", "sl")
		test("es-419", "es-419")
		test("es", "es")
		test("sw", "sw")
		test("sv", "sv")
		test("ta", "ta")
		test("te", "te")
		test("th", "th")
		test("tr", "tr")
		test("uk", "uk")
		test("vi", "vi")
		test("zu", "zu")
	})

}
