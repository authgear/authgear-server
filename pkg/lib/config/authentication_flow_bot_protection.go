package config

var _ = Schema.Add("AuthenticationFlowBotProtection", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["mode"],
	"properties": {
		"mode": { "$ref": "#/$defs/BotProtectionRiskMode" }
	}
}
`)

type AuthenticationFlowBotProtection struct {
	Mode BotProtectionRiskMode `json:"mode,omitempty"`
}
