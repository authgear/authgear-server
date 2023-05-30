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

type WhatsappTemplateType string

const (
	WhatsappTemplateTypeAuthentication WhatsappTemplateType = "authentication"
)

var _ = Schema.Add("WhatsappTemplateType", `
{
	"type": "string",
	"enum": ["authentication"]
}
`)

type WhatsappTemplateConfig struct {
	Name      string               `json:"name"`
	Type      WhatsappTemplateType `json:"type"`
	Languages []string             `json:"languages"`
}

var _ = Schema.Add("WhatsappTemplateConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "minLength": 1 },
		"type": { "$ref": "#/$defs/WhatsappTemplateType" },
		"languages": {
			"type": "array",
			"items": {
				"type": "string",
				"minLength": 1
			}
		}
	},
	"required": ["name", "type", "languages"]
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
				"required": ["templates"]
			}
		}
	]
}
`)

type WhatsappConfig struct {
	APIType   WhatsappAPIType          `json:"api_type,omitempty"`
	Templates *WhatsappTemplatesConfig `json:"templates,omitempty"`
}

func (c *WhatsappConfig) NullableFields() []string {
	return []string{"Templates"}
}
