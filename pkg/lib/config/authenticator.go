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
		"expiry": { "$ref": "#/$defs/PasswordExpiryConfig" },
		"force_change": { "type": "boolean" },
		"ratelimit": { "$ref": "#/$defs/PasswordRatelimitConfig" }
	}
}
`)

type AuthenticatorPasswordConfig struct {
	Policy               *PasswordPolicyConfig    `json:"policy,omitempty"`
	Expiry               *PasswordExpiryConfig    `json:"expiry,omitempty"`
	ForceChange          *bool                    `json:"force_change,omitempty"`
	Deprecated_Ratelimit *PasswordRatelimitConfig `json:"ratelimit,omitempty"`
}

func (c *AuthenticatorPasswordConfig) SetDefaults() {
	if c.ForceChange == nil {
		c.ForceChange = newBool(true)
	}

	c.Deprecated_Ratelimit = nil
}

var _ = Schema.Add("PasswordPolicyConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"min_length": { "type": "integer", "minimum": 1 },
		"uppercase_required": { "type": "boolean" },
		"lowercase_required": { "type": "boolean" },
		"alphabet_required": { "type": "boolean" },
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
	AlphabetRequired      bool         `json:"alphabet_required,omitempty"`
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

var _ = Schema.Add("PasswordExpiryConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"force_change": { "$ref": "#/$defs/PasswordExpiryForceChangeConfig" }
	}
}
`)

var _ = Schema.Add("PasswordExpiryForceChangeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"duration_since_last_update": { "$ref": "#/$defs/DurationString" }
	},
	"if": { "properties": { "enabled": { "const": true } }, "required": ["enabled"] },
	"then": { "required": ["duration_since_last_update"] }
}
`)

/**
Example config of password expiry

NOTE: Currently only force_change is supported. The other 2 cases are planned in later phase.

```
password:
	# The 3 cases can be turned on individually.
	# However, the precedence is deny_login > force_change > prompt_change.
	expiry: # "expiration" is American English, while "expiry" is British English. But we have been using "expiry" in the config. So let's stick with it.
		# In this case, the authflow will result in a dead end, with an error telling that the password is expired and the login is denied.
		deny_login:
			enabled: true
			duration_since_last_update: 2160h
		# In this case, the authflow will enter the change_password step (The existing one). To proceed, the end-user must change the password.
		force_change:
				enabled: true
      duration_since_last_update: 1440h
    # In this case, the authflow will enter the change_password step. But this step now supports taking an input to skip the password update.
    prompt_change:
      enabled: true
      duration_since_last_update: 720h
```
**/

type PasswordExpiryConfig struct {
	ForceChange *PasswordExpiryForceChangeConfig `json:"force_change,omitempty"`
}

func (c *PasswordExpiryConfig) SetDefaults() {
	if c.ForceChange == nil {
		c.ForceChange = &PasswordExpiryForceChangeConfig{}
	}
}

type PasswordExpiryForceChangeConfig struct {
	Enabled                 bool           `json:"enabled,omitempty"`
	DurationSinceLastUpdate DurationString `json:"duration_since_last_update,omitempty"`
}

func (c *PasswordExpiryForceChangeConfig) IsEnabled() bool {
	sinceLastUpdate, sinceLastUpdateIsValid := c.DurationSinceLastUpdate.MaybeDuration()
	return c.Enabled && sinceLastUpdateIsValid && sinceLastUpdate > 0
}

var _ = Schema.Add("PasswordRatelimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"failed_attempt": { "$ref": "#/$defs/PasswordFailedAttemptConfig" }
	}
}
`)

// PasswordRatelimitConfig is deprecated
type PasswordRatelimitConfig struct {
	FailedAttempt *PasswordFailedAttemptConfig `json:"failed_attempt,omitempty"`
}

var _ = Schema.Add("PasswordFailedAttemptConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"size": {
			"type": "integer",
			"minimum": 1
		},
		"reset_period": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type PasswordFailedAttemptConfig struct {
	Size        int            `json:"size,omitempty"`
	ResetPeriod DurationString `json:"reset_period,omitempty"`
}

func (c *PasswordFailedAttemptConfig) SetDefaults() {
	if c.Size == 0 {
		c.Size = 10
	}
	if c.ResetPeriod == "" {
		c.ResetPeriod = "1m"
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

func (m *AuthenticatorPhoneOTPMode) IsSMSEnabled() bool {
	return *m == AuthenticatorPhoneOTPModeWhatsappSMS ||
		*m == AuthenticatorPhoneOTPModeSMSOnly
}

var _ = Schema.Add("AuthenticatorOOBValidPeriods", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"link": { "$ref": "#/$defs/DurationString" },
		"code": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type AuthenticatorOOBValidPeriods struct {
	Link DurationString `json:"link,omitempty"`
	Code DurationString `json:"code,omitempty"`
}

var _ = Schema.Add("AuthenticatorOOBSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"phone_otp_mode": { "$ref": "#/$defs/AuthenticatorPhoneOTPMode" },
		"code_valid_period": { "$ref": "#/$defs/DurationString" },
		"valid_periods": { "$ref": "#/$defs/AuthenticatorOOBValidPeriods" }
	}
}
`)

type AuthenticatorOOBSMSConfig struct {
	Maximum                    *int                          `json:"maximum,omitempty"`
	PhoneOTPMode               AuthenticatorPhoneOTPMode     `json:"phone_otp_mode,omitempty"`
	Deprecated_CodeValidPeriod DurationString                `json:"code_valid_period,omitempty"`
	ValidPeriods               *AuthenticatorOOBValidPeriods `json:"valid_periods,omitempty"`
}

func (c *AuthenticatorOOBSMSConfig) SetDefaults() {
	if c.PhoneOTPMode == "" {
		c.PhoneOTPMode = AuthenticatorPhoneOTPModeWhatsappSMS
	}
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
	if c.Deprecated_CodeValidPeriod == "" {
		c.Deprecated_CodeValidPeriod = DurationString("300s")
	}
	if c.ValidPeriods.Code == "" {
		c.ValidPeriods.Code = c.Deprecated_CodeValidPeriod
	}
	if c.ValidPeriods.Link == "" {
		c.ValidPeriods.Link = DurationString("20m")
	}
	// See https://github.com/authgear/authgear-server/issues/4297
	// Remove deprecated fields
	c.Deprecated_CodeValidPeriod = ""
}

var _ = Schema.Add("AuthenticatorOOBEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"email_otp_mode": { "$ref": "#/$defs/AuthenticatorEmailOTPMode" },
		"code_valid_period": { "$ref": "#/$defs/DurationString" },
		"valid_periods": { "$ref": "#/$defs/AuthenticatorOOBValidPeriods" }
	}
}
`)

type AuthenticatorOOBEmailConfig struct {
	Maximum                    *int                          `json:"maximum,omitempty"`
	EmailOTPMode               AuthenticatorEmailOTPMode     `json:"email_otp_mode,omitempty"`
	Deprecated_CodeValidPeriod DurationString                `json:"code_valid_period,omitempty"`
	ValidPeriods               *AuthenticatorOOBValidPeriods `json:"valid_periods,omitempty"`
}

var _ = Schema.Add("AuthenticatorEmailOTPMode", `
{
	"type": "string",
	"enum": ["code", "login_link"]
}
`)

type AuthenticatorEmailOTPMode string

const (
	AuthenticatorEmailOTPModeCodeOnly      AuthenticatorEmailOTPMode = "code"
	AuthenticatorEmailOTPModeLoginLinkOnly AuthenticatorEmailOTPMode = "login_link"
)

func (m *AuthenticatorEmailOTPMode) IsCodeEnabled() bool {
	return *m == AuthenticatorEmailOTPModeCodeOnly
}

func (m *AuthenticatorEmailOTPMode) IsLoginLinkEnabled() bool {
	return *m == AuthenticatorEmailOTPModeLoginLinkOnly
}

func (c *AuthenticatorOOBEmailConfig) SetDefaults() {
	if c.EmailOTPMode == "" {
		c.EmailOTPMode = AuthenticatorEmailOTPModeCodeOnly
	}
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
	switch c.EmailOTPMode {
	case AuthenticatorEmailOTPModeCodeOnly:
		if c.Deprecated_CodeValidPeriod == "" {
			c.Deprecated_CodeValidPeriod = DurationString("300s")
		}
		if c.ValidPeriods.Link == "" {
			c.ValidPeriods.Link = DurationString("20m")
		}
		if c.ValidPeriods.Code == "" {
			c.ValidPeriods.Code = c.Deprecated_CodeValidPeriod
		}
	case AuthenticatorEmailOTPModeLoginLinkOnly:
		if c.Deprecated_CodeValidPeriod == "" {
			c.Deprecated_CodeValidPeriod = DurationString("20m")
		}
		if c.ValidPeriods.Link == "" {
			c.ValidPeriods.Link = c.Deprecated_CodeValidPeriod
		}
		if c.ValidPeriods.Code == "" {
			c.ValidPeriods.Code = DurationString("300s")
		}
	default:
		panic("unknown email otp mode")
	}
	// See https://github.com/authgear/authgear-server/issues/3524
	// Remove deprecated fields
	c.Deprecated_CodeValidPeriod = ""
}
