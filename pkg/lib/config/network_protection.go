package config

var _ = Schema.Add("NetworkProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ip_blocklist": { "$ref": "#/$defs/NetworkIPBlocklistConfig" }
	}
}
`)

var _ = Schema.Add("NetworkIPBlocklistConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"cidrs": {
			"type": "array",
			"items": { "type": "string", "format": "x_cidr" }
		},
		"country_codes": {
			"type": "array",
			"items": { "type": "string" }
		}
	}
}
`)

type NetworkProtectionConfig struct {
	IPBlocklist *NetworkIPBlocklistConfig `json:"ip_blocklist,omitempty"`
}

type NetworkIPBlocklistConfig struct {
	CIDRs        []string `json:"cidrs,omitempty"`
	CountryCodes []string `json:"country_codes,omitempty"`
}
