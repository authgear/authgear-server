package config

var _ = Schema.Add("NetworkProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ip_filter": { "$ref": "#/$defs/IPFilterConfig" }
	}
}
`)

var _ = Schema.Add("IPFilterAction", `
{
	"type": "string",
	"enum": ["allow", "deny"]
}
`)

var _ = Schema.Add("IPFilterSource", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"cidrs": {
			"type": "array",
			"items": { "type": "string", "format": "x_cidr" }
		},
		"geo_location_codes": {
			"type": "array",
			"items": { "type": "string", "minLength": 2, "maxLength": 2 }
		}
	}
}
`)

var _ = Schema.Add("IPFilterRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string" },
		"action": { "$ref": "#/$defs/IPFilterAction" },
		"source": { "$ref": "#/$defs/IPFilterSource" }
	},
	"required": ["action", "source"]
}
`)

var _ = Schema.Add("IPFilterConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"default_action": { "$ref": "#/$defs/IPFilterAction" },
		"rules": {
			"type": "array",
			"items": { "$ref": "#/$defs/IPFilterRule" }
		}
	}
}
`)

type NetworkProtectionConfig struct {
	IPFilter *IPFilterConfig `json:"ip_filter,omitempty"`
}

type IPFilterConfig struct {
	DefaultAction IPFilterAction  `json:"default_action,omitempty"`
	Rules         []*IPFilterRule `json:"rules,omitempty"`
}

func (c *IPFilterConfig) SetDefaults() {
	if c.DefaultAction == "" {
		c.DefaultAction = IPFilterActionAllow
	}
}

type IPFilterAction string

const (
	IPFilterActionAllow IPFilterAction = "allow"
	IPFilterActionDeny  IPFilterAction = "deny"
)

type IPFilterRule struct {
	Name   string         `json:"name"`
	Action IPFilterAction `json:"action"`
	Source IPFilterSource `json:"source"`
}

type IPFilterSource struct {
	CIDRs            []string `json:"cidrs,omitempty"`
	GeoLocationCodes []string `json:"geo_location_codes,omitempty"`
}
