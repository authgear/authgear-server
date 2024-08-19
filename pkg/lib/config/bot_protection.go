package config

var _ = Schema.Add("BotProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
    "provider": { "$ref": "#/$defs/BotProtectionProvider" },
		"requirements": { "$ref": "#/$defs/BotProtectionRequirements" }
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
		"site_key": { "type": "string", "minLength": 1 }
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

var _ = Schema.Add("BotProtectionRequirements", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"signup_or_login": { "$ref": "#/$defs/BotProtectionRequirementsObject" },
		"account_recovery": { "$ref": "#/$defs/BotProtectionRequirementsObject" },
		"password": { "$ref": "#/$defs/BotProtectionRequirementsObject" },
		"oob_otp_email": { "$ref": "#/$defs/BotProtectionRequirementsObject" },
		"oob_otp_sms": { "$ref": "#/$defs/BotProtectionRequirementsObject" }
	}
}
`)

var _ = Schema.Add("BotProtectionRequirementsObject", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"mode": { "$ref": "#/$defs/BotProtectionRiskMode" }
	},
	"required": ["mode"]
}
`)

var _ = Schema.Add("BotProtectionRiskMode", `
{
	"type": "string",
	"enum": ["never", "always"]
}
`)

type BotProtectionConfig struct {
	Enabled      bool                       `json:"enabled,omitempty"`
	Provider     *BotProtectionProvider     `json:"provider,omitempty" nullable:"true"`
	Requirements *BotProtectionRequirements `json:"requirements,omitempty" nullable:"true"`
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
	SiteKey string                    `json:"site_key,omitempty"` // only for cloudflare, recaptchav2
}

type BotProtectionProviderType string

const (
	BotProtectionProviderTypeCloudflare  BotProtectionProviderType = "cloudflare"
	BotProtectionProviderTypeRecaptchaV2 BotProtectionProviderType = "recaptchav2"
)

type BotProtectionRequirements struct {
	SignupOrLogin   *BotProtectionRequirementsObject `json:"signup_or_login,omitempty"`
	AccountRecovery *BotProtectionRequirementsObject `json:"account_recovery,omitempty"`
	Password        *BotProtectionRequirementsObject `json:"password,omitempty"`
	OOBOTPEmail     *BotProtectionRequirementsObject `json:"oob_otp_email,omitempty"`
	OOBOTPSMS       *BotProtectionRequirementsObject `json:"oob_otp_sms,omitempty"`
}

type BotProtectionRequirementsObject struct {
	Mode BotProtectionRiskMode `json:"mode,omitempty"`
}

// NOTE: If you add any new BotProtectionRiskMode, please make corresponding changes in GetStricterBotProtectionRiskMode too
type BotProtectionRiskMode string

const (
	BotProtectionRiskModeNever  BotProtectionRiskMode = "never"
	BotProtectionRiskModeAlways BotProtectionRiskMode = "always"
)

func GetStricterBotProtectionRiskMode(rmA BotProtectionRiskMode, rmB BotProtectionRiskMode) BotProtectionRiskMode {
	if rmA == BotProtectionRiskModeAlways || rmB == BotProtectionRiskModeAlways {
		return BotProtectionRiskModeAlways
	}

	// Add more risk modes here

	return BotProtectionRiskModeNever
}
