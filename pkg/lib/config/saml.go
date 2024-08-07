package config

var _ = Schema.Add("SAMLConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"service_providers": {
			"type": "array",
			"items": { "$ref": "#/$defs/SAMLServiceProviderConfig" }
		}
	}
}
`)

var _ = Schema.Add("SAMLServiceProviderConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"id": { "type": "string" }
	},
	"required": ["id"]
}
`)

type SAMLConfig struct {
	SAMLServiceProviders []*SAMLServiceProviderConfig `json:"service_providers,omitempty"`
}

type SAMLServiceProviderConfig struct {
	ID string `json:"id,omitempty"`
}
