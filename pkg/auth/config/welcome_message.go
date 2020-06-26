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
		"email_message": { "$ref": "#/$defs/EmailMessageConfig" },
		"destination": { "$ref": "#/$defs/WelcomeMessageDestination" }
	}
}
`)

type WelcomeMessageConfig struct {
	Enabled      bool                      `json:"enabled,omitempty"`
	EmailMessage EmailMessageConfig        `json:"email_message,omitempty"`
	Destination  WelcomeMessageDestination `json:"destination,omitempty"`
}

func (c *WelcomeMessageConfig) SetDefaults() {
	if c.Destination == "" {
		c.Destination = WelcomeMessageDestinationFirst
	}
	if c.EmailMessage["subject"] == "" {
		c.EmailMessage["subject"] = "Welcome!"
	}
}
