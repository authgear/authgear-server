package config

var _ = Schema.Add("AuthenticationFlowCaptcha", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["mode"],
	"properties": {
		"mode": { "type": "string", "enum": ["never", "always"] }
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
				"required": ["provider"],
				"properties": {
					"provider": {
						"type": "object",
						"additionalProperties": false,
						"required": ["alias"],
						"properties": {
							"alias": { "type": "string" }
						}
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowCaptchaProvider", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alias": { "type": "string" }
	}
}
`)

type AuthenticationFlowCaptchaProvider struct {
	Alias string `json:"alias,omitempty"`
}
type AuthenticationFlowCaptcha struct {
	Mode     AuthenticationFlowCaptchaMode      `json:"mode,omitempty"`
	Provider *AuthenticationFlowCaptchaProvider `json:"provider,omitempty" nullable:"true"`
}

func (c *AuthenticationFlowCaptcha) SetDefaults() {
	c.Mode = AuthenticationFlowCaptchaModeNever
}

type AuthenticationFlowCaptchaMode string

const (
	AuthenticationFlowCaptchaModeNever  AuthenticationFlowCaptchaMode = "never"
	AuthenticationFlowCaptchaModeAlways AuthenticationFlowCaptchaMode = "always"
)
