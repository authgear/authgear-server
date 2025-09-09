package config

type WhatsappOnPremisesTemplatesConfig struct {
	OTP WhatsappOnPremisesOTPTemplateConfig `json:"otp"`
}

var _ = SecretConfigSchema.Add("WhatsappOnPremisesTemplatesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"otp": { "$ref": "#/$defs/WhatsappOnPremisesOTPTemplateConfig" }
	},
	"required": ["otp"]
}
`)

type WhatsappOnPremisesTemplateType string

const (
	WhatsappOnPremisesTemplateTypeAuthentication WhatsappOnPremisesTemplateType = "authentication"
)

var _ = SecretConfigSchema.Add("WhatsappOnPremisesTemplateType", `
{
	"type": "string",
	"enum": ["authentication"]
}
`)

type WhatsappOnPremisesOTPTemplateConfig struct {
	Name      string                         `json:"name"`
	Type      WhatsappOnPremisesTemplateType `json:"type"`
	Namespace string                         `json:"namespace,omitempty"`
	Languages []string                       `json:"languages"`
}

var _ = SecretConfigSchema.Add("WhatsappOnPremisesOTPTemplateConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "minLength": 1 },
		"type": { "$ref": "#/$defs/WhatsappOnPremisesTemplateType" },
		"namespace": { "type": "string", "minLength": 1 },
		"languages": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/WhatsappTemplateLanguage"
			},
			"minItems": 1
		}
	},
	"required": ["name", "type", "languages"]
}
`)

var _ = SecretConfigSchema.Add("WhatsappOnPremisesCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_endpoint": { "type": "string", "minLength": 1 },
		"username": { "type": "string", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 },
		"templates": { "$ref": "#/$defs/WhatsappOnPremisesTemplatesConfig" }
	},
	"required": ["api_endpoint", "username", "password", "templates"]
}
`)

type WhatsappOnPremisesCredentials struct {
	APIEndpoint string                             `json:"api_endpoint"`
	Username    string                             `json:"username"`
	Password    string                             `json:"password"`
	Templates   *WhatsappOnPremisesTemplatesConfig `json:"templates"`
}

func (c *WhatsappOnPremisesCredentials) SensitiveStrings() []string {
	return []string{
		c.Username,
		c.Password,
	}
}

// WhatsappCloudAPIAuthenticationTemplateType is NOT an official term.
// Officially, they just call authentication templates "authentication templates".
// See https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/
// The variants are
//   - One-Tap Autofill Authentication Templates
//     https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/#one-tap-autofill-authentication-templates
//   - Copy Code Authentication Templates
//     https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/#copy-code-authentication-templates
//   - Zero-Tap Authentication Templates
//     https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/#zero-tap-authentication-templates
//
// We only support the Copy Code variant.
type WhatsappCloudAPIAuthenticationTemplateType string

const (
	WhatsappCloudAPIAuthenticationTemplateTypeCopyCodeButton WhatsappCloudAPIAuthenticationTemplateType = "copy_code_button"
)

type WhatsappCloudAPIAuthenticationTemplateConfig struct {
	Type           WhatsappCloudAPIAuthenticationTemplateType                  `json:"type"`
	CopyCodeButton *WhatsappCloudAPIAuthenticationTemplateCopyCodeButtonConfig `json:"copy_code_button"`
}

type WhatsappCloudAPIAuthenticationTemplateCopyCodeButtonConfig struct {
	Name      string   `json:"name"`
	Languages []string `json:"languages"`
}

type WhatsappCloudAPIWebhook struct {
	VerifyToken string `json:"verify_token"`
}

type WhatsappCloudAPICredentials struct {
	PhoneNumberID                string                                        `json:"phone_number_id"`
	AccessToken                  string                                        `json:"access_token"`
	AuthenticationTemplateConfig *WhatsappCloudAPIAuthenticationTemplateConfig `json:"authentication_template"`
	Webhook                      *WhatsappCloudAPIWebhook                      `json:"webhook,omitempty"`
}

func (c *WhatsappCloudAPICredentials) SensitiveStrings() []string {
	return []string{
		c.PhoneNumberID,
		c.AccessToken,
	}
}

// The list of template languages of On-Premises API can be found at
// https://developers.facebook.com/docs/whatsapp/api/messages/message-templates
// The list of template languages of Cloud CPI can be found at
// https://developers.facebook.com/docs/whatsapp/business-management-api/message-templates/supported-languages
// This list was compiled on 2025-03-26. The above 2 lists are the same on that day.
var _ = SecretConfigSchema.Add("WhatsappTemplateLanguage", `
{
	"type": "string",
	"enum": [
		"af",
		"sq",
		"ar",
		"ar_EG",
		"ar_AE",
		"ar_LB",
		"ar_MA",
		"ar_QA",
		"az",
		"be_BY",
		"bn",
		"bn_IN",
		"bg",
		"ca",
		"zh_CN",
		"zh_HK",
		"zh_TW",
		"hr",
		"cs",
		"da",
		"prs_AF",
		"nl",
		"nl_BE",
		"en",
		"en_GB",
		"en_US",
		"en_AE",
		"en_AU",
		"en_CA",
		"en_GH",
		"en_IE",
		"en_IN",
		"en_JM",
		"en_MY",
		"en_NZ",
		"en_QA",
		"en_SG",
		"en_UG",
		"en_ZA",
		"et",
		"fil",
		"fi",
		"fr",
		"fr_BE",
		"fr_CA",
		"fr_CH",
		"fr_CI",
		"fr_MA",
		"ka",
		"de",
		"de_AT",
		"de_CH",
		"el",
		"gu",
		"ha",
		"he",
		"hi",
		"hu",
		"id",
		"ga",
		"it",
		"ja",
		"kn",
		"kk",
		"rw_RW",
		"ko",
		"ky_KG",
		"lo",
		"lv",
		"lt",
		"mk",
		"ms",
		"ml",
		"mr",
		"nb",
		"ps_AF",
		"fa",
		"pl",
		"pt_BR",
		"pt_PT",
		"pa",
		"ro",
		"ru",
		"sr",
		"si_LK",
		"sk",
		"sl",
		"es",
		"es_AR",
		"es_CL",
		"es_CO",
		"es_CR",
		"es_DO",
		"es_EC",
		"es_HN",
		"es_MX",
		"es_PA",
		"es_PE",
		"es_ES",
		"es_UY",
		"sw",
		"sv",
		"ta",
		"te",
		"th",
		"th",
		"tr",
		"uk",
		"ur",
		"uz",
		"vi",
		"zu"
	]
}
`)

var _ = SecretConfigSchema.Add("WhatsappCloudAPIAuthenticationTemplateCopyCodeButtonConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "minLength": 1 },
		"languages": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/WhatsappTemplateLanguage"
			},
			"minItems": 1
		}
	},
	"required": ["name", "languages"]
}
`)

var _ = SecretConfigSchema.Add("WhatsappCloudAPIAuthenticationTemplateConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "const": "copy_code_button" },
		"copy_code_button": { "$ref": "#/$defs/WhatsappCloudAPIAuthenticationTemplateCopyCodeButtonConfig" }
	},
	"required": ["type", "copy_code_button"]
}
`)

var _ = SecretConfigSchema.Add("WhatsappCloudAPIWebhook", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"verify_token": { "type": "string", "minLength": 1 }
	},
	"required": ["verify_token"]
}
`)

var _ = SecretConfigSchema.Add("WhatsappCloudAPICredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"phone_number_id": { "type": "string", "minLength": 1 },
		"access_token": { "type": "string", "minLength": 1 },
		"authentication_template": { "$ref": "#/$defs/WhatsappCloudAPIAuthenticationTemplateConfig" },
		"webhook": { "$ref": "#/$defs/WhatsappCloudAPIWebhook" }
	},
	"required": ["phone_number_id", "access_token", "authentication_template"]
}
`)
