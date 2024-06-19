package config

var _ = Schema.Add("AuthenticationFlowCaptcha", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"required": { "type": "boolean" }
	}
}
`)

type AuthenticationFlowCaptcha struct {
	Required *bool `json:"required,omitempty"`
}
