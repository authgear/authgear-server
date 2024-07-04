package config

var _ = Schema.Add("AuthenticationFlowBotProtection", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["mode"],
	"properties": {
		"mode": { "type": "string", "enum": ["never", "always"] },
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
	Mode     AuthenticationFlowBotProtectionMode      `json:"mode,omitempty"`
	Provider *AuthenticationFlowBotProtectionProvider `json:"provider,omitempty"`
}

type AuthenticationFlowBotProtectionMode string

const (
	AuthenticationFlowBotProtectionModeNever  AuthenticationFlowBotProtectionMode = "never"
	AuthenticationFlowBotProtectionModeAlways AuthenticationFlowBotProtectionMode = "always"
)
