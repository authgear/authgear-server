package intl

import "fmt"

// ResolveRecaptchaV2 resolves language tag to RecaptchaV2-supported language tags
// ref https://developers.google.com/recaptcha/docs/language
func ResolveRecaptchaV2(lang string) string {
	if !AvailableLanguagesMap[lang] {
		panic(fmt.Errorf("unsupported language: %s", lang))
	}
	switch lang {
	case "es-ES":
		return "es"
	default:
		return lang
	}
}

// https://developers.google.com/recaptcha/docs/language as of 2024-07-23
var RecaptchaV2Languages = []string{
	"af",
	"sq",
	"am",
	"ar",
	"hy",
	"as",
	"az",
	"eu",
	"be",
	"bn",
	"bg",
	"my",
	"ca",
	"zh-HK",
	"zh-CN",
	"zh-TW",
	"hr",
	"cs",
	"da",
	"nl",
	"en-GB",
	"en",
	"et",
	"fil",
	"fi",
	"fr",
	"fr-CA",
	"gl",
	"ka",
	"de",
	"de-AT",
	"de-CH",
	"el",
	"gu",
	"iw",
	"hi",
	"hu",
	"is",
	"id",
	"ga",
	"it",
	"ja",
	"kn",
	"kk",
	"km",
	"ko",
	"ky",
	"lo",
	"lv",
	"lt",
	"mk",
	"ms",
	"ml",
	"mr",
	"mn",
	"ne",
	"no",
	"or",
	"fa",
	"pl",
	"pt",
	"pt-BR",
	"pt-PT",
	"ro",
	"ru",
	"sr",
	"si",
	"sk",
	"sl",
	"es",
	"es-419",
	"sw",
	"sv",
	"ta",
	"te",
	"th",
	"tr",
	"uk",
	"ur",
	"uz",
	"vi",
	"cy",
	"zu",
}
