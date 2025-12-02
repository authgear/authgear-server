package config

var _ = Schema.Add("ProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ip_blocklist": { "$ref": "#/$defs/IPBlocklistConfig" }
	}
}
`)

var _ = Schema.Add("IPBlocklistConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"cidrs": {
			"type": "array",
			"items": { "type": "string" }
		},
		"country_codes": {
			"type": "array",
			"items": { "type": "string" }
		}
	}
}
`)

type ProtectionConfig struct {
	IPBlocklist *IPBlocklistConfig `json:"ip_blocklist,omitempty"`
}

type IPBlocklistConfig struct {
	CIDRs        []string `json:"cidrs,omitempty"`
	CountryCodes []string `json:"country_codes,omitempty"`
}
