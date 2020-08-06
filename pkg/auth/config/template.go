package config

var _ = Schema.Add("TemplateConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"items": { "type": "array", "items": { "$ref": "#/$defs/TemplateItem" } }
	}
}
`)

type TemplateConfig struct {
	Items []TemplateItem `json:"items,omitempty"`
}

var _ = Schema.Add("TemplateItemType", `{ "type": "string" }`)

type TemplateItemType string

var _ = Schema.Add("TemplateItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "$ref": "#/$defs/TemplateItemType" },
		"language_tag": { "type": "string" },
		"uri": { "type": "string" }
	},
	"required": ["type", "uri"]
}
`)

type TemplateItem struct {
	Type        TemplateItemType `json:"type,omitempty"`
	LanguageTag string           `json:"language_tag,omitempty"`
	URI         string           `json:"uri,omitempty"`
}
