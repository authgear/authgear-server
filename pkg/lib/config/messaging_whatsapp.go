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

type WhatsappTemplateConfig struct {
	Name       string                           `json:"name"`
	Languages  []string                         `json:"languages"`
	Components *WhatsappTemplateComponentConfig `json:"components"`
}

var _ = Schema.Add("WhatsappTemplateConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "minLength": 1 },
		"languages": {
			"type": "array",
			"items": {
				"type": "string",
				"minLength": 1
			}
		},
		"components": {
			"$ref": "#/$defs/WhatsappTemplateComponentConfig"
		}
	},
	"required": ["name", "languages", "components"]
}
`)

type WhatsappTemplateComponentConfig struct {
	Header *WhatsappTemplateComponentParameter `json:"header,omitempty"`
	Body   *WhatsappTemplateComponentParameter `json:"body,omitempty"`
}

var _ = Schema.Add("WhatsappTemplateComponentConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"header": {
			"$ref": "#/$defs/WhatsappTemplateComponentParameter"
		},
		"body": {
			"$ref": "#/$defs/WhatsappTemplateComponentParameter"
		}
	}
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
	}
}
`)

var _ = Schema.Add("WhatsappConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_type": { "$ref": "#/$defs/WhatsappAPIType" },
		"api_endpoint": { "type": "string", "minLength": 1 },
		"templates": { "$ref": "#/$defs/WhatsappTemplatesConfig" }
	},
	"required": ["api_type", "templates"],
	"allOf": [
		{
			"if": { "properties": { "api_type": { "const": "on-premises" } } },
			"then": {
				"required": ["api_endpoint"]
			}
		}
	]
}
`)

type WhatsappConfig struct {
	APIType     WhatsappAPIType          `json:"api_type"`
	APIEndpoint *string                  `json:"api_endpoint,omitempty"`
	Templates   *WhatsappTemplatesConfig `json:"templates"`
}
