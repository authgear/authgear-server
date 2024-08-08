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
		"id": { "type": "string" },
		"entity_id": { "type": "string" }
	},
	"required": ["id", "entity_id"]
}
`)

type SAMLConfig struct {
	SAMLServiceProviders []*SAMLServiceProviderConfig `json:"service_providers,omitempty"`
}

func (c *SAMLConfig) ResolveProvider(id string) (*SAMLServiceProviderConfig, bool) {
	for _, sp := range c.SAMLServiceProviders {
		if sp.ID == id {
			return sp, true
		}
	}
	return nil, false
}

type SAMLServiceProviderConfig struct {
	ID       string `json:"id,omitempty"`
	EntityID string `json:"entity_id,omitempty"`
}
