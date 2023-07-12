package config

var _ = Schema.Add("WelcomeMessageDestination", `
{
	"type": "string",
	"enum": ["first", "all"]
}
`)

type WelcomeMessageDestination string

const (
	WelcomeMessageDestinationFirst WelcomeMessageDestination = "first"
	WelcomeMessageDestinationAll   WelcomeMessageDestination = "all"
)

var _ = Schema.Add("WelcomeMessageConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"destination": { "$ref": "#/$defs/WelcomeMessageDestination" }
	}
}
`)

type WelcomeMessageConfig struct {
	Enabled     bool                      `json:"enabled,omitempty"`
	Destination WelcomeMessageDestination `json:"destination,omitempty"`
}

func (c *WelcomeMessageConfig) SetDefaults() {
	if c.Destination == "" {
		c.Destination = WelcomeMessageDestinationFirst
	}
}
