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
