package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"country_calling_code": { "$ref": "#/$defs/UICountryCallingCodeConfig" },
		"dark_theme_disabled": { "type": "boolean" },
		"default_client_uri": { "type": "string", "format": "uri" },
		"default_redirect_uri": { "type": "string", "format": "uri" },
		"default_post_logout_redirect_uri": { "type": "string", "format": "uri" }
	}
}
`)

type UIConfig struct {
	CountryCallingCode *UICountryCallingCodeConfig `json:"country_calling_code,omitempty"`
	DarkThemeDisabled  bool                        `json:"dark_theme_disabled,omitempty"`
	// client_uri to use when client_id is absent.
	DefaultClientURI string `json:"default_client_uri,omitempty"`
	// redirect_uri to use when client_id is absent.
	DefaultRedirectURI string `json:"default_redirect_uri,omitempty"`
	// post_logout_redirect_uri to use when client_id is absent.
	DefaultPostLogoutRedirectURI string `json:"default_post_logout_redirect_uri,omitempty"`
}

var _ = Schema.Add("UICountryCallingCodeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"allowlist": { "type": "array", "items": { "type": "string" }, "minItems": 1 },
		"pinned_list": { "type": "array", "items": { "type": "string" } }
	}
}
`)

type UICountryCallingCodeConfig struct {
	AllowList  []string `json:"allowlist,omitempty"`
	PinnedList []string `json:"pinned_list,omitempty"`
}

func (c *UICountryCallingCodeConfig) SetDefaults() {
	if c.AllowList == nil {
		c.AllowList = phone.CountryCallingCodes
	}
}

// NOTE: Pinned list has order, goes before allow list
// while allow list follows default order.
// Country code in either pinned / allow list is counted as active
func (c *UICountryCallingCodeConfig) GetActiveCountryCodes() []string {
	isCodePinned := make(map[string]bool)
	for _, code := range c.PinnedList {
		isCodePinned[code] = true
	}

	activeCountryCodes := []string{}
	activeCountryCodes = append(activeCountryCodes, c.PinnedList...)
	for _, code := range c.AllowList {
		if !isCodePinned[code] {
			activeCountryCodes = append(activeCountryCodes, code)
		}
	}

	return activeCountryCodes
}
