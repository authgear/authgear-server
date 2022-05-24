package config

var _ = Schema.Add("AuthenticatorConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"password": { "$ref": "#/$defs/AuthenticatorPasswordConfig" },
		"totp": { "$ref": "#/$defs/AuthenticatorTOTPConfig" },
		"oob_otp": { "$ref": "#/$defs/AuthenticatorOOBConfig" }
	}
}
`)

type AuthenticatorConfig struct {
	Password *AuthenticatorPasswordConfig `json:"password,omitempty"`
	TOTP     *AuthenticatorTOTPConfig     `json:"totp,omitempty"`
	OOB      *AuthenticatorOOBConfig      `json:"oob_otp,omitempty"`
}

var _ = Schema.Add("AuthenticatorPasswordConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"policy": { "$ref": "#/$defs/PasswordPolicyConfig" },
		"force_change": { "type": "boolean" }
	}
}
`)

type AuthenticatorPasswordConfig struct {
	Policy      *PasswordPolicyConfig `json:"policy,omitempty"`
	ForceChange *bool                 `json:"force_change,omitempty"`
}

func (c *AuthenticatorPasswordConfig) SetDefaults() {
	if c.ForceChange == nil {
		c.ForceChange = newBool(true)
	}
}

var _ = Schema.Add("PasswordPolicyConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"min_length": { "type": "integer", "minimum": 1 },
		"uppercase_required": { "type": "boolean" },
		"lowercase_required": { "type": "boolean" },
		"digit_required": { "type": "boolean" },
		"symbol_required": { "type": "boolean" },
		"minimum_guessable_level": { "type": "integer" },
		"excluded_keywords": { "type": "array", "items": { "type": "string" } },
		"history_size": { "type": "integer" },
		"history_days": { "$ref": "#/$defs/DurationDays" }
	}
}
`)

type PasswordPolicyConfig struct {
	MinLength             *int         `json:"min_length,omitempty"`
	UppercaseRequired     bool         `json:"uppercase_required,omitempty"`
	LowercaseRequired     bool         `json:"lowercase_required,omitempty"`
	DigitRequired         bool         `json:"digit_required,omitempty"`
	SymbolRequired        bool         `json:"symbol_required,omitempty"`
	MinimumGuessableLevel int          `json:"minimum_guessable_level,omitempty"`
	ExcludedKeywords      []string     `json:"excluded_keywords,omitempty"`
	HistorySize           int          `json:"history_size,omitempty"`
	HistoryDays           DurationDays `json:"history_days,omitempty"`
}

func (c *PasswordPolicyConfig) IsEnabled() bool {
	return c.HistorySize > 0 || c.HistoryDays > 0
}

func (c *PasswordPolicyConfig) SetDefaults() {
	if c.MinLength == nil {
		c.MinLength = newInt(8)
	}
}

var _ = Schema.Add("AuthenticatorTOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" }
	}
}
`)

type AuthenticatorTOTPConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

func (c *AuthenticatorTOTPConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
}

var _ = Schema.Add("AuthenticatorOOBConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms": { "$ref": "#/$defs/AuthenticatorOOBSMSConfig" },
		"email": { "$ref": "#/$defs/AuthenticatorOOBEmailConfig" }
	}
}
`)

type AuthenticatorOOBConfig struct {
	SMS   *AuthenticatorOOBSMSConfig   `json:"sms,omitempty"`
	Email *AuthenticatorOOBEmailConfig `json:"email,omitempty"`
}

var _ = Schema.Add("AuthenticatorPhoneOTPMode", `
{
	"type": "string",
	"enum": ["sms", "whatsapp_sms", "whatsapp"]
}
`)

type AuthenticatorPhoneOTPMode string

const (
	AuthenticatorPhoneOTPModeSMSOnly      AuthenticatorPhoneOTPMode = "sms"
	AuthenticatorPhoneOTPModeWhatsappSMS  AuthenticatorPhoneOTPMode = "whatsapp_sms"
	AuthenticatorPhoneOTPModeWhatsappOnly AuthenticatorPhoneOTPMode = "whatsapp"
)

func (m *AuthenticatorPhoneOTPMode) IsWhatsappEnabled() bool {
	return *m == AuthenticatorPhoneOTPModeWhatsappSMS ||
		*m == AuthenticatorPhoneOTPModeWhatsappOnly
}

var _ = Schema.Add("AuthenticatorOOBSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"phone_otp_mode": { "$ref": "#/$defs/AuthenticatorPhoneOTPMode" }
	}
}
`)

type AuthenticatorOOBSMSConfig struct {
	Maximum      *int                      `json:"maximum,omitempty"`
	PhoneOTPMode AuthenticatorPhoneOTPMode `json:"phone_otp_mode,omitempty"`
}

func (c *AuthenticatorOOBSMSConfig) SetDefaults() {
	if c.PhoneOTPMode == "" {
		c.PhoneOTPMode = AuthenticatorPhoneOTPModeSMSOnly
	}
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
}

var _ = Schema.Add("AuthenticatorOOBEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" }
	}
}
`)

type AuthenticatorOOBEmailConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

func (c *AuthenticatorOOBEmailConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
}
