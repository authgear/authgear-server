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
		"default_post_logout_redirect_uri": { "type": "string", "format": "uri" }
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
}

var _ = Schema.Add("PhoneInputConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"allowlist": { "type": "array", "items": { "$ref": "#/$defs/ISO31661Alpha2" }, "minItems": 1 },
		"pinned_list": { "type": "array", "items": { "$ref": "#/$defs/ISO31661Alpha2" } }
	}
}
`)

var _ = Schema.Add("ISO31661Alpha2", phone.JSONSchemaString)

type PhoneInputConfig struct {
	AllowList  []string `json:"allowlist,omitempty"`
	PinnedList []string `json:"pinned_list,omitempty"`
}

func (c *PhoneInputConfig) SetDefaults() {
	if c.AllowList == nil {
		c.AllowList = phone.AllAlpha2
	}
}

// GetCountries returns a list of countries, with pinned countries go first,
// followed by allowed countries.
func (c *PhoneInputConfig) GetCountries() []phone.Country {
	var countries []phone.Country

	isPinned := make(map[string]bool)
	for _, alpha2 := range c.PinnedList {
		isPinned[alpha2] = true
	}

	// Insert pinned country in order.
	for _, alpha2 := range c.PinnedList {
		country, ok := phone.GetCountryByAlpha2(alpha2)
		if ok {
			countries = append(countries, country)
		}
	}

	// Insert allowed country in order.
	for _, alpha2 := range c.AllowList {
		pinned := isPinned[alpha2]
		if pinned {
			continue
		}
		country, ok := phone.GetCountryByAlpha2(alpha2)
		if ok {
			countries = append(countries, country)
		}
	}

	return countries
}
