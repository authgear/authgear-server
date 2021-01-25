package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"country_calling_code": { "$ref": "#/$defs/UICountryCallingCodeConfig" },
		"dark_theme_disabled": { "type": "boolean" }
	}
}
`)

type UIConfig struct {
	CountryCallingCode *UICountryCallingCodeConfig `json:"country_calling_code,omitempty"`
	DarkThemeDisabled  bool                        `json:"dark_theme_disabled,omitempty"`
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
