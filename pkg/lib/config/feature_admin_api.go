package config

var _ = FeatureConfigSchema.Add("AdminAPIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"create_session_enabled": { "type": "boolean" },
		"user_import_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"user_export_usage": { "$ref": "#/$defs/UsageLimitConfig" }
	}
}
`)

type AdminAPIFeatureConfig struct {
	CreateSessionEnabled *bool `json:"create_session_enabled,omitempty"`
	// UserImportUsage is the usage limit on user import API, measured by number of imported users.
	UserImportUsage *UsageLimitConfig `json:"user_import_usage,omitempty"`
	// UserExportUsage is the usage limit on user export API, measured by number of export requests.
	UserExportUsage *UsageLimitConfig `json:"user_export_usage,omitempty"`
}

func (c *AdminAPIFeatureConfig) SetDefaults() {
	if c.CreateSessionEnabled == nil {
		c.CreateSessionEnabled = newBool(false)
	}
	if c.UserImportUsage.Enabled == nil {
		c.UserImportUsage = &UsageLimitConfig{
			Enabled: newBool(false),
		}
	}
	if c.UserExportUsage.Enabled == nil {
		c.UserExportUsage = &UsageLimitConfig{
			Enabled: newBool(false),
		}
	}
}
