package config

var _ = Schema.Add("AccountMigrationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"hook": { "$ref": "#/$defs/AccountMigrationHookConfig" },
		"proof_of_phone_number_verification": { "$ref": "#/$defs/ProofOfPhoneNumberVerificationConfig" }
	}
}
`)

type AccountMigrationConfig struct {
	Hook                           *AccountMigrationHookConfig           `json:"hook,omitempty"`
	ProofOfPhoneNumberVerification *ProofOfPhoneNumberVerificationConfig `json:"proof_of_phone_number_verification,omitempty"`
}

var _ = Schema.Add("AccountMigrationHookConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"url": { "type": "string", "format": "x_hook_uri" },
		"timeout": { "type": "integer" }
	}
}
`)

type AccountMigrationHookConfig struct {
	URL     string          `json:"url,omitempty"`
	Timeout DurationSeconds `json:"timeout,omitempty"`
}

func (c *AccountMigrationHookConfig) SetDefaults() {
	if c.Timeout == 0 {
		c.Timeout = DurationSeconds(5)
	}
}
