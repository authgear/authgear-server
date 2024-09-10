package main

var allowedLegitTranslationKeys []string = []string{
	"app.name",
	"customer-support-link",
	"privacy-policy-link",
	"terms-of-service-link",
}

var goTemplateFalstPositives []string = []string{
	"widget",
	"page-frame-content",
	"page-content",
	"dialog-attr",
	"dialog-controller-str",
	"dialog-content",
	"dialog-close-btn",
	"authflowv2/__error_account",
	"page-navbar",
	"__forgot_password_alternative",
	"__settings_profile_item",
	"__settings_profile_date_item",
	"__settings_profile_address_item",
	"__settings_profile_locale_item",
	"__settings_gender_edit_custom_gender_input",
}

var allowedVariableKeys []string = []string{
	"$label",
	"$labelKey",
}

var AllowedKeys map[string]struct{}

func init() {
	AllowedKeys = make(map[string]struct{})
	for _, k := range allowedLegitTranslationKeys {
		AllowedKeys[k] = struct{}{}
	}
	for _, k := range goTemplateFalstPositives {
		AllowedKeys[k] = struct{}{}
	}
	for _, k := range allowedVariableKeys {
		AllowedKeys[k] = struct{}{}
	}
}

func IsSpecialCase(translationKey string) bool {
	_, ok := AllowedKeys[translationKey]
	return ok
}
