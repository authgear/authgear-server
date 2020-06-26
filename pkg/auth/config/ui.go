package config

import "github.com/skygeario/skygear-server/pkg/core/phone"

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"custom_css": { "type": "string" },
		"country_calling_code": { "$ref": "#/$defs/UICountryCallingCodeConfig" }
	}
}
`)

type UIConfig struct {
	CustomCSS          string                      `json:"custom_css,omitempty"`
	CountryCallingCode *UICountryCallingCodeConfig `json:"country_calling_code,omitempty"`
}

var _ = Schema.Add("UICountryCallingCodeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"values": { "type": "array", "items": { "type": "string" } },
		"default": { "type": "string" }
	}
}
`)

type UICountryCallingCodeConfig struct {
	Values  []string `json:"values,omitempty"`
	Default string   `json:"default,omitempty"`
}

func (c *UICountryCallingCodeConfig) SetDefaults() {
	if c.Values == nil {
		c.Values = phone.CountryCallingCodes
	}
	if c.Default == "" {
		c.Default = c.Values[0]
	}
}
