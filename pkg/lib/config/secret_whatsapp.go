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

// ref: https://developers.facebook.com/docs/whatsapp/api/messages/message-templates
var _ = SecretConfigSchema.Add("WhatsappOnPremisesTemplateLanguage", `
{
	"type": "string",
	"enum": [
		"af", "sq", "ar", "az", "bn",
		"bg", "ca", "zh_CN", "zh_HK", "zh_TW",
		"hr", "cs", "da", "nl", "en",
		"en_GB", "en_US", "et", "fil", "fi",
		"fr", "ka", "de", "el", "gu",
		"ha", "he", "hi", "hu", "id",
		"ga", "it", "ja", "kn", "kk",
		"rw_RW", "ko", "ky_KG", "lo", "lv",
		"lt", "mk", "ms", "ml", "mr",
		"nb", "fa", "pl", "pt_BR", "pt_PT",
		"pa", "ro", "ru", "sr", "sk",
		"sl", "es", "es_AR", "es_ES", "es_MX",
		"sw", "sv", "ta", "te", "th",
		"tr", "uk", "ur", "uz", "vi",
		"zu"
	]
}
`)

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
				"$ref": "#/$defs/WhatsappOnPremisesTemplateLanguage",
				"minLength": 1
			}
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
