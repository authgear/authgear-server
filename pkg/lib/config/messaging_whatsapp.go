package config

var _ = Schema.Add("WhatsappAPIType", `
{
	"type": "string",
	"enum": ["on-premises"]
}
`)

type WhatsappAPIType string

const (
	WhatsappAPITypeOnPremises WhatsappAPIType = "on-premises"
)

type WhatsappTemplatesConfig struct {
	OTP WhatsappTemplateConfig `json:"otp"`
}

var _ = Schema.Add("WhatsappTemplatesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"otp": { "$ref": "#/$defs/WhatsappTemplateConfig" }
	},
	"required": ["otp"]
}
`)

var _ = Schema.Add("WhatsappOnPremTemplatesConfig", `
{
	"allOf": [
		{ "$ref": "#/$defs/WhatsappTemplatesConfig" }
	],
	"properties": {
		"otp": { "$ref": "#/$defs/WhatsappOnPremTemplateConfig" }
	}
}
`)

type WhatsappTemplateType string

const (
	WhatsappTemplateTypeAuthentication WhatsappTemplateType = "authentication"
)

// ref: https://developers.facebook.com/docs/whatsapp/api/messages/message-templates
var _ = Schema.Add("WhatsappTemplateLanguage", `
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

var _ = Schema.Add("WhatsappTemplateType", `
{
	"type": "string",
	"enum": ["authentication"]
}
`)

type WhatsappTemplateConfig struct {
	Name      string               `json:"name"`
	Type      WhatsappTemplateType `json:"type"`
	Namespace string               `json:"namespace,omitempty"`
	Languages []string             `json:"languages"`
}

var _ = Schema.Add("WhatsappTemplateConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "minLength": 1 },
		"type": { "$ref": "#/$defs/WhatsappTemplateType" },
		"namespace": { "type": "string", "minLength": 1 },
		"languages": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/WhatsappTemplateLanguage",
				"minLength": 1
			}
		}
	},
	"required": ["name", "type", "languages"]
}
`)

var _ = Schema.Add("WhatsappOnPremTemplateConfig", `
{
	"type": "object",
	"allOf": [
		{ "$ref": "#/$defs/WhatsappTemplateConfig" },
		{
			"required": ["namespace"]
		}
	]
}
`)

type WhatsappTemplateComponentParameter struct {
	Parameters []string `json:"parameters,omitempty"`
}

var _ = Schema.Add("WhatsappTemplateComponentParameter", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"parameters": {
			"type": "array",
			"items": { "type": "string", "minLength": 1 }
		}
	},
	"required": ["parameters"]
}
`)

var _ = Schema.Add("WhatsappConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_type": { "$ref": "#/$defs/WhatsappAPIType" },
		"templates": { "$ref": "#/$defs/WhatsappTemplatesConfig" }
	},
	"allOf": [
		{
			"if": {
        "properties": { "api_type": { "const": "on-premises" } },
        "required": ["api_type"]
      },
			"then": {
        "properties": {
					"templates": { "$ref": "#/$defs/WhatsappOnPremTemplatesConfig" }
				},
				"required": ["templates"]
			}
		}
	]
}
`)

type WhatsappConfig struct {
	APIType   WhatsappAPIType          `json:"api_type,omitempty"`
	Templates *WhatsappTemplatesConfig `json:"templates,omitempty" nullable:"true"`
}
