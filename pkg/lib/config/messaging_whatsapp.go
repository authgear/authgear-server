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

func (c *WhatsappTemplatesConfig) IsNullable() bool {
	return true
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
	},
	"required": ["parameters"]
}
`)

var _ = Schema.Add("WhatsappConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"api_type": { "$ref": "#/$defs/WhatsappAPIType" },
		"api_endpoint": { "type": "string", "minLength": 1 },
		"templates": { "$ref": "#/$defs/WhatsappTemplatesConfig" }
	},
	"allOf": [
		{
			"if": {
				"properties": { "enabled": { "enum": [true] } },
        "required": ["enabled"]
			},
			"then": {
				"required": ["api_type", "templates"]
			}
		},
		{
			"if": {
        "properties": { "api_type": { "const": "on-premises" } },
        "required": ["api_type"]
      },
			"then": {
				"required": ["api_endpoint"]
			}
		}
	]
}
`)

type WhatsappConfig struct {
	Enabled     bool                     `json:"enabled"`
	APIType     WhatsappAPIType          `json:"api_type,omitempty"`
	APIEndpoint string                   `json:"api_endpoint,omitempty"`
	Templates   *WhatsappTemplatesConfig `json:"templates,omitempty"`
}

func (c *WhatsappConfig) NullableFields() []string {
	return []string{"Templates"}
}
