package config

var _ = Schema.Add("AuthenticationFlowBotProtection", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["mode"],
	"properties": {
		"mode": { "$ref": "#/$defs/BotProtectionRiskMode" },
		"provider": { "$ref": "#/$defs/AuthenticationFlowBotProtectionProvider" }
	},
	"allOf": [
		{
			"if": {
				"required": ["mode"],
				"properties": {
					"mode": {
						"enum": ["always"]
					}
				}
			},
			"then": {
				"required": ["provider"]
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowBotProtectionProvider", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["type"],
	"properties": {
		"type": { "type": "string", "enum": ["cloudflare", "recaptchav2"] }
	}
}
`)

type AuthenticationFlowBotProtectionProvider struct {
	Type BotProtectionProviderType `json:"type,omitempty"`
}
type AuthenticationFlowBotProtection struct {
	Mode     BotProtectionRiskMode                    `json:"mode,omitempty"`
	Provider *AuthenticationFlowBotProtectionProvider `json:"provider,omitempty"`
}
