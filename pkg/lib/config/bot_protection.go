package config

var _ = Schema.Add("BotProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
    "provider": { "$ref": "#/$defs/BotProtectionProvider" }
	}
}
`)

var _ = Schema.Add("BotProtectionProvider", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["type"],
	"properties": {
		"type": { "type": "string", "enum": ["cloudflare", "recaptchav2"] },
		"site_key": { "type": "string" }
	},
	"allOf": [
		{
			"if": {
				"properties": {
					"type": {
						"enum": ["cloudflare", "recaptchav2"]
					}
				},
				"required": ["type"]
			},
			"then": {
				"required": ["site_key"]
			}
		}
	]
}
`)

type BotProtectionConfig struct {
	Enabled  bool                   `json:"enabled,omitempty"`
	Provider *BotProtectionProvider `json:"provider,omitempty" nullable:"true"`
}

func (c *BotProtectionConfig) IsEnabled() bool {
	return c != nil && c.Enabled && c.Provider != nil && c.Provider.Type != "" && c.Provider.SiteKey != ""
}
func (c *BotProtectionConfig) GetProviderType() BotProtectionProviderType {
	if !c.IsEnabled() {
		return ""
	}
	return c.Provider.Type
}
func (c *BotProtectionConfig) GetSiteKey() string {
	if !c.IsEnabled() {
		return ""
	}
	return c.Provider.SiteKey
}

type BotProtectionProvider struct {
	Type    BotProtectionProviderType `json:"type,omitempty"`
	SiteKey string                    `json:"site_key,omitempty"` // only for some cloudflare, recaptchav2
}

type BotProtectionProviderType string

const (
	BotProtectionProviderTypeCloudflare  BotProtectionProviderType = "cloudflare"
	BotProtectionProviderTypeRecaptchaV2 BotProtectionProviderType = "recaptchav2"
)
