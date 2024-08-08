package intl

import "fmt"

// ResolveCloudflareTurnstile resolves language tag to Cloudflare-Turnstile-supported language tags
// ref https://developers.cloudflare.com/turnstile/reference/supported-languages/
func ResolveCloudflareTurnstile(lang string) string {
	if _, ok := AvailableLanguagesMap[lang]; !ok {
		panic(fmt.Errorf("unsupported language: %s", lang))
	}
	switch lang {
	case "fil":
		// "fil" is the standardized form of "tl"
		// ref https://en.wikipedia.org/wiki/Tagalog_language
		return "tl-PH"
	case "be":
		// Belarusian not supported by Cloudflare Turnstile
		// return Russian - another official language in Belarus instead
		// ref https://en.wikipedia.org/wiki/Belarusian_language
		return "ru"
	case "eu":
		// Basque not supported by Cloudflare Turnstile
		// return French - another official language in Basque instead
		// ref https://en.wikipedia.org/wiki/Basque_Country_(greater_region)
		return "fr"
	case "bn":
		// Bengali not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Bangladesh instead
		// ref https://en.wikipedia.org/wiki/Bangladesh
		return "en"
	case "ka":
		// Georgian not supported by Cloudflare Turnstile
		// return Russian - the closest country to Georgia which Turnstile support instead
		// ref https://en.wikipedia.org/wiki/Georgia_(country)
		return "ru"
	case "hy":
		// Armenian not supported by Cloudflare Turnstile
		// return Russian - "the best known foreign language" in Armenia instead
		// ref https://en.wikipedia.org/wiki/Languages_of_Armenia
		return "ru"
	case "kn":
		// Kannada not supported by Cloudflare Turnstile
		// return Hindi - another official language in India instead
		// ref https://en.wikipedia.org/wiki/Languages_with_legal_status_in_India
		return "hi"
	case "ta":
		// Tamil not supported by Cloudflare Turnstile
		// return Hindi - another official language in India instead
		// ref https://en.wikipedia.org/wiki/Languages_with_legal_status_in_India
		return "hi"
	case "km":
		// Khmer not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Cambodia instead
		// ref https://en.wikipedia.org/wiki/Demographics_of_Cambodia#Languages
		return "en"
	case "lv":
		// Latvian not supported by Cloudflare Turnstile
		// return Russian - the closest country to Latvia which Turnstile support instead
		// ref https://en.wikipedia.org/wiki/Latvia
		return "ru"
	case "si":
		// Sinhala not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Sri Lanka instead
		// ref https://en.wikipedia.org/wiki/Sri_Lanka
		return "en"
	case "zu":
		// Zulu not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in South Africa instead
		// ref https://en.wikipedia.org/wiki/Languages_of_South_Africa
		return "en"
	case "ne":
		// Nepali not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Nepal instead
		// ref https://en.wikipedia.org/wiki/Languages_of_Nepal
		return "en"
	case "sw":
		// Swahili not supported by Cloudflare Turnstile
		// return English - another official in Tanzania
		// ref https://en.wikipedia.org/wiki/Tanzania
		return "en"
	case "mn":
		// Mongolian not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Mongolia instead
		// Fun fact - Since 1990, English has quickly supplanted Russian as the most popular foreign language in Mongolia
		// ref https://en.wikipedia.org/wiki/Mongolian_language
		return "en"
	case "gl":
		// Galician not supported by Cloudflare Turnstile
		// return Spanish - the official language of Spain - the Sovereign of Galicia instead
		// ref https://en.wikipedia.org/wiki/Galicia_(Spain)
		return "es"
	case "pt-PT":
		// Portuguese (Portugal) not supported by Cloudflare Turnstile
		// return Portuguese (Brazil) instead
		return "pt"
	case "no":
		// Norwegian not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Norway instead
		// ref https://en.wikipedia.org/wiki/Languages_of_Norway
		return "en"
	case "ky":
		// Kyrgyz not supported by Cloudflare Turnstile
		// return Russian - another official language in Kyrgyzstan
		// ref https://en.wikipedia.org/wiki/Kyrgyzstan
		return "ru"
	case "az":
		// Azerbaijani not supported by Cloudflare Turnstile
		// return Russian - another official language in Azerbaijan
		// ref https://en.wikipedia.org/wiki/Azerbaijan
		return "ru"
	case "ml":
		// Malayalam not supported by Cloudflare Turnstile
		// return Hindi - another official language in India instead
		// ref https://en.wikipedia.org/wiki/Languages_with_legal_status_in_India
		return "hi"
	case "ca":
		// Catalan not supported by Cloudflare Turnstile
		// return Spanish - the official language of Spain - the Sovereign of Andorra instead
		// ref https://en.wikipedia.org/wiki/Catalan_language
		return "es"
	case "te":
		// Telugu not supported by Cloudflare Turnstile
		// return Hindi - another official language in India instead
		// ref https://en.wikipedia.org/wiki/Languages_with_legal_status_in_India
		return "hi"
	case "zh-HK":
		// Chinese (Hong Kong) not supported by Cloudflare Turnstile
		return "zh-TW"
		// return Chinese (Taiwan) - also traditional Chinese instead
	case "is":
		// Icelandic not supported by Cloudflare Turnstile
		// return English - mandatory language to study in Iceland instead
		// ref https://en.wikipedia.org/wiki/Languages_of_Iceland
		return "en"
	case "et":
		// Estonian not supported by Cloudflare Turnstile
		// return English - most widely spoken foreign language in Estonia today
		// ref https://en.wikipedia.org/wiki/Estonia
		return "en"
	case "af":
		// Afrikaans not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in South Africa instead
		// ref https://en.wikipedia.org/wiki/Languages_of_South_Africa
		return "en"
	case "lo":
		// Lao not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Laos instead
		// ref https://en.wikipedia.org/wiki/Laos
		return "en"
	case "mr":
		// Marathi not supported by Cloudflare Turnstile
		// return Hindi - another official language in India instead
		// ref https://en.wikipedia.org/wiki/Languages_with_legal_status_in_India
		return "hi"
	case "my":
		// Burmese not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Myanmar instead
		// ref https://en.wikipedia.org/wiki/Myanmar
		return "en"
	case "mk":
		// Macedonian not supported by Cloudflare Turnstile
		// return Turkish - a minority language in North Macedonia
		// ref https://en.wikipedia.org/wiki/Languages_of_North_Macedonia
		return "tr"
	case "am":
		// Amharic not supported by Cloudflare Turnstile
		// return English - a recognized foreign language in Ethiopia instead
		// ref https://en.wikipedia.org/wiki/Ethiopia
		return "en"
	case "es-419":
		// Spanish appropriate for the Latin America and Caribbean region
		// is not supported by Cloudflare Turnstile
		// return Spanish instead
		return "es"
	// both authgear- & turnstile-supported
	default:
		return lang
	}
}

// https://developers.cloudflare.com/turnstile/reference/supported-languages/ as of 2024-07-23
var CloudflareTurnstileLanguages = []string{
	"ar-EG",
	"ar",
	"bg-BG",
	"bg",
	"zh-CN",
	"zh",
	"zh-TW",
	"hr-HR",
	"hr",
	"cs-CZ",
	"cs",
	"da-DK",
	"da",
	"nl-NL",
	"nl",
	"en-US",
	"en",
	"fa-IR",
	"fa",
	"fi-FI",
	"fi",
	"fr-FR",
	"fr",
	"de-DE",
	"de",
	"el-GR",
	"el",
	"he-IL",
	"he",
	"hi-IN",
	"hi",
	"hu-HU",
	"hu",
	"id-ID",
	"id",
	"it-IT",
	"it",
	"ja-JP",
	"ja",
	"tlh",
	"ko-KR",
	"ko",
	"lt-LT",
	"lt",
	"ms-MY",
	"ms",
	"nb-NO",
	"nb",
	"pl-PL",
	"pl",
	"pt-BR",
	"pt",
	"ro-RO",
	"ro",
	"ru-RU",
	"ru",
	"sr-BA",
	"sr",
	"sk-SK",
	"sk",
	"sl-SI",
	"sl",
	"es-ES",
	"es",
	"sv-SE",
	"sv",
	"tl-PH",
	"tl",
	"th-TH",
	"th",
	"tr-TR",
	"tr",
	"uk-UA",
	"uk",
	"vi-VN",
	"vi",
}
