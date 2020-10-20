package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"country_calling_code": { "$ref": "#/$defs/UICountryCallingCodeConfig" }
	}
}
`)

type UIConfig struct {
	CountryCallingCode *UICountryCallingCodeConfig `json:"country_calling_code,omitempty"`
}

var _ = Schema.Add("UICountryCallingCodeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"allow_list": { "type": "array", "items": { "type": "string" }, "minItems": 1 },
		"pinned_list": { "type": "array", "items": { "type": "string" }, "minItems": 1 }
	}
}
`)

type UICountryCallingCodeConfig struct {
	AllowList  []string `json:"allow_list,omitempty"`
	PinnedList []string `json:"pinned_list,omitempty"`
}

func (c *UICountryCallingCodeConfig) SetDefaults() {
	if c.AllowList == nil {
		c.AllowList = phone.CountryCallingCodes
	}
	if c.PinnedList == nil {
		c.PinnedList = []string{c.AllowList[0]}
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
