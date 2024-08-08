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

func GetStrictestAuthFlowBotProtection(bpList ...*AuthenticationFlowBotProtection) *AuthenticationFlowBotProtection {
	if len(bpList) == 0 {
		panic("unexpected empty bpList")
	}
	out := &AuthenticationFlowBotProtection{
		Mode: bpList[0].Mode,
	}

	for _, currBP := range bpList {
		out.Mode = GetStricterBotProtectionRiskMode(out.Mode, currBP.Mode)
	}

	return out
}
