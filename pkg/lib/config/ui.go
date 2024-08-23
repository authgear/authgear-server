package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"signup_login_flow_enabled": { "type": "boolean" },
		"phone_input": { "$ref": "#/$defs/PhoneInputConfig" },
		"dark_theme_disabled": { "type": "boolean" },
		"light_theme_disabled": { "type": "boolean" },
		"watermark_disabled": { "type": "boolean" },
		"direct_access_disabled": { "type": "boolean" },
		"default_client_uri": { "type": "string", "format": "uri" },
		"default_redirect_uri": { "type": "string", "format": "uri" },
		"default_branding_page_uri": { "type": "string", "format": "uri" },
		"default_post_logout_redirect_uri": { "type": "string", "format": "uri" },
		"authentication_disabled": { "type": "boolean" },
		"settings_disabled": { "type": "boolean" },
		"implementation": {
			"type": "string",
			"enum": ["interaction", "authflow", "authflowv2"]
		},
		"settings_implementation": {
			"type": "string",
			"enum": ["v1", "v2"]
		},
		"forgot_password": { "$ref": "#/$defs/UIForgotPasswordConfig" },
		"authentication_flow": { "$ref": "#/$defs/UIAuthenticationFlowConfig" }
	}
}
`)

type UIConfig struct {
	SignupLoginFlowEnabled bool              `json:"signup_login_flow_enabled,omitempty"`
	PhoneInput             *PhoneInputConfig `json:"phone_input,omitempty"`
	DarkThemeDisabled      bool              `json:"dark_theme_disabled,omitempty"`
	LightThemeDisabled     bool              `json:"light_theme_disabled,omitempty"`
	WatermarkDisabled      bool              `json:"watermark_disabled,omitempty"`
	DirectAccessDisabled   bool              `json:"direct_access_disabled,omitempty"`
	// client_uri to use when client_id is absent.
	DefaultClientURI string `json:"default_client_uri,omitempty"`
	// branding_page_uri to use when client_id is absent.
	DefaultBrandingPageURI string `json:"default_branding_page_uri,omitempty"`
	// redirect_uri to use when client_id is absent.
	DefaultRedirectURI string `json:"default_redirect_uri,omitempty"`
	// post_logout_redirect_uri to use when client_id is absent.
	DefaultPostLogoutRedirectURI string `json:"default_post_logout_redirect_uri,omitempty"`
	// NOTE: Internal use only, use authentication_disabled to disable auth-ui when custom ui is used
	AuthenticationDisabled bool `json:"authentication_disabled,omitempty"`
	SettingsDisabled       bool `json:"settings_disabled,omitempty"`
	// Implementation is a temporary flag to switch between authflow and interaction.
	Implementation UIImplementation `json:"implementation,omitempty"`
	// SettingImplementation is a temporary flag to switch between setting ui v1 and v2.
	SettingsImplementation SettingsUIImplementation `json:"settings_implementation,omitempty"`
	// ForgotPassword is the config for the default auth ui
	ForgotPassword *UIForgotPasswordConfig `json:"forgot_password,omitempty"`
	// AuthenticationFlow configures ui behaviour of authentication flow under default auth ui
	AuthenticationFlow *UIAuthenticationFlowConfig `json:"authentication_flow,omitempty"`
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
	UIImplementationAuthflowV2  UIImplementation = "authflowv2"
)

type SettingsUIImplementation string

const (
	SettingsUIImplementationDefault SettingsUIImplementation = ""
	SettingsUIImplementationV1      SettingsUIImplementation = "v1"
	SettingsUIImplementationV2      SettingsUIImplementation = "v2"
)

var _ = Schema.Add("UIForgotPasswordConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"phone": { "type": "array", "items": { "$ref": "#/$defs/AccountRecoveryChannel" } },
		"email": { "type": "array", "items": { "$ref": "#/$defs/AccountRecoveryChannel" } }
	}
}
`)

type UIForgotPasswordConfig struct {
	Phone []*AccountRecoveryChannel `json:"phone,omitempty"`
	Email []*AccountRecoveryChannel `json:"email,omitempty"`
}

func (c *UIForgotPasswordConfig) SetDefaults() {
	if c.Phone == nil {
		c.Phone = []*AccountRecoveryChannel{
			{
				Channel: AccountRecoveryCodeChannelSMS,
				OTPForm: AccountRecoveryCodeFormCode,
			},
		}
	}

	if c.Email == nil {
		c.Email = []*AccountRecoveryChannel{
			{
				Channel: AccountRecoveryCodeChannelEmail,
				OTPForm: AccountRecoveryCodeFormLink,
			},
		}
	}
}

var _ = Schema.Add("UIAuthenticationFlowConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"groups": { "type": "array", "items": { "$ref": "#/$defs/UIAuthenticationFlowGroup" } }
	}
}
`)

var _ = Schema.Add("UIAuthenticationFlowGroup", `
{
	"type": "object",
	"additionalProperties": false,
	"required": [
		"name",
		"flows"
	],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"flows" : { "type": "array", "items": { "$ref": "#/$defs/UIAuthenticationFlowGroupFlow" } }
	}
}
`)

var _ = Schema.Add("UIAuthenticationFlowGroupFlow", `
{
	"type": "object",
	"additionalProperties": false,
	"required": [
		"type",
		"name"
	],
	"properties": {
		"type": {
			"type": "string",
			"enum": [
				"signup",
				"promote",
				"login",
				"signup_login",
				"reauth",
				"account_recovery"
			]
		},
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" }
	}
}
`)

type UIAuthenticationFlowConfig struct {
	Groups []*UIAuthenticationFlowGroup `json:"groups,omitempty"`
}

type UIAuthenticationFlowGroupFlow struct {
	Type AuthenticationFlowType `json:"type"`
	Name string                 `json:"name"`
}

type UIAuthenticationFlowGroup struct {
	Name  string                           `json:"name"`
	Flows []*UIAuthenticationFlowGroupFlow `json:"flows,omitempty"`
}
