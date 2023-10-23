package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"phone_input": { "$ref": "#/$defs/PhoneInputConfig" },
		"dark_theme_disabled": { "type": "boolean" },
		"watermark_disabled": { "type": "boolean" },
		"default_client_uri": { "type": "string", "format": "uri" },
		"default_redirect_uri": { "type": "string", "format": "uri" },
		"default_post_logout_redirect_uri": { "type": "string", "format": "uri" },
		"authentication_disabled": { "type": "boolean" },
		"settings_disabled": { "type": "boolean" },
		"implementation": {
			"type": "string",
			"enum": ["interaction", "authflow"]
		}
	}
}
`)

type UIConfig struct {
	PhoneInput        *PhoneInputConfig `json:"phone_input,omitempty"`
	DarkThemeDisabled bool              `json:"dark_theme_disabled,omitempty"`
	WatermarkDisabled bool              `json:"watermark_disabled,omitempty"`
	// client_uri to use when client_id is absent.
	DefaultClientURI string `json:"default_client_uri,omitempty"`
	// redirect_uri to use when client_id is absent.
	DefaultRedirectURI string `json:"default_redirect_uri,omitempty"`
	// post_logout_redirect_uri to use when client_id is absent.
	DefaultPostLogoutRedirectURI string `json:"default_post_logout_redirect_uri,omitempty"`
	// NOTE: Internal use only, use authentication_disabled to disable auth-ui when custom ui is used
	AuthenticationDisabled bool `json:"authentication_disabled,omitempty"`
	SettingsDisabled       bool `json:"settings_disabled,omitempty"`
	// Implementation is a temporary flag to switch between authflow and interaction.
	Implementation UIImplementation `json:"implementation,omitempty"`
}

var _ = Schema.Add("PhoneInputConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"allowlist": { "type": "array", "items": { "$ref": "#/$defs/ISO31661Alpha2" }, "minItems": 1 },
		"pinned_list": { "type": "array", "items": { "$ref": "#/$defs/ISO31661Alpha2" } },
		"preselect_by_ip_disabled": { "type": "boolean" }
	}
}
`)

var _ = Schema.Add("ISO31661Alpha2", phone.JSONSchemaString)

type PhoneInputConfig struct {
	AllowList             []string `json:"allowlist,omitempty"`
	PinnedList            []string `json:"pinned_list,omitempty"`
	PreselectByIPDisabled bool     `json:"preselect_by_ip_disabled,omitempty"`
}

func (c *PhoneInputConfig) SetDefaults() {
	if c.AllowList == nil {
		c.AllowList = phone.AllAlpha2
	}
}

type UIImplementation string

const (
	UIImplementationDefault     UIImplementation = ""
	UIImplementationInteraction UIImplementation = "interaction"
	UIImplementationAuthflow    UIImplementation = "authflow"
)
