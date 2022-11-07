package config

var _ = Schema.Add("RateLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type RateLimitConfig struct {
	Disabled *bool `json:"disabled,omitempty"`
}

func (c *RateLimitConfig) SetDefaults() {
	if c.Disabled == nil {
		c.Disabled = newBool(false)
	}
}
