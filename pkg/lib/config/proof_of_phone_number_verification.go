package config

var _ = Schema.Add("ProofOfPhoneNumberVerificationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"hook": { "$ref": "#/$defs/ProofOfPhoneNumberVerificationHookConfig" }
	}
}
`)

type ProofOfPhoneNumberVerificationConfig struct {
	Hook *ProofOfPhoneNumberVerificationHookConfig `json:"hook,omitempty"`
}

var _ = Schema.Add("ProofOfPhoneNumberVerificationHookConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"url": { "type": "string", "format": "x_hook_uri" },
		"timeout": { "type": "integer" }
	}
}
`)

type ProofOfPhoneNumberVerificationHookConfig struct {
	URL     string          `json:"url,omitempty"`
	Timeout DurationSeconds `json:"timeout,omitempty"`
}

func (c *ProofOfPhoneNumberVerificationHookConfig) SetDefaults() {
	if c.Timeout == 0 {
		c.Timeout = DurationSeconds(5)
	}
}
