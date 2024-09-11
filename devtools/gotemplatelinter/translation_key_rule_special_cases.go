package main

var allowedLegitTranslationKeys []string = []string{
	"app.name",
	"customer-support-link",
	"privacy-policy-link",
	"terms-of-service-link",
}

var AllowedKeys map[string]struct{}

func init() {
	AllowedKeys = make(map[string]struct{})
	for _, k := range allowedLegitTranslationKeys {
		AllowedKeys[k] = struct{}{}
	}
}

func IsSpecialCase(translationKey string) bool {
	_, ok := AllowedKeys[translationKey]
	return ok
}
