package config

var _ = Schema.Add("AuthenticationFlowObjectCaptchaConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["enabled"],
	"properties": {
		"enabled": { "type": "boolean" }
	}
}
`)

type AuthenticationFlowObjectCaptchaConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

var _ = Schema.Add("AuthenticationFlowCaptcha", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["required"],
	"properties": {
		"required": { "type": "boolean" }
	}
}
`)

type AuthenticationFlowCaptcha struct {
	Required *bool `json:"required,omitempty"`
}
