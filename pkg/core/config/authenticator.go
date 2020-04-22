package config

//go:generate msgp -tests=false
type AuthenticatorConfiguration struct {
	Password     *AuthenticatorPasswordConfiguration     `json:"password,omitempty" yaml:"password" msg:"password" default_zero_value:"true"`
	TOTP         *AuthenticatorTOTPConfiguration         `json:"totp,omitempty" yaml:"totp" msg:"totp" default_zero_value:"true"`
	OOB          *AuthenticatorOOBConfiguration          `json:"oob_otp,omitempty" yaml:"oob_otp" msg:"oob_otp" default_zero_value:"true"`
	BearerToken  *AuthenticatorBearerTokenConfiguration  `json:"bearer_token,omitempty" yaml:"bearer_token" msg:"bearer_token" default_zero_value:"true"`
	RecoveryCode *AuthenticatorRecoveryCodeConfiguration `json:"recovery_code,omitempty" yaml:"recovery_code" msg:"recovery_code" default_zero_value:"true"`
}

type AuthenticatorPasswordConfiguration struct {
	Policy *PasswordPolicyConfiguration `json:"policy,omitempty" yaml:"policy" msg:"policy" default_zero_value:"true"`
}

type PasswordPolicyConfiguration struct {
	MinLength             int      `json:"min_length,omitempty" yaml:"min_length" msg:"min_length"`
	UppercaseRequired     bool     `json:"uppercase_required,omitempty" yaml:"uppercase_required" msg:"uppercase_required"`
	LowercaseRequired     bool     `json:"lowercase_required,omitempty" yaml:"lowercase_required" msg:"lowercase_required"`
	DigitRequired         bool     `json:"digit_required,omitempty" yaml:"digit_required" msg:"digit_required"`
	SymbolRequired        bool     `json:"symbol_required,omitempty" yaml:"symbol_required" msg:"symbol_required"`
	MinimumGuessableLevel int      `json:"minimum_guessable_level,omitempty" yaml:"minimum_guessable_level" msg:"minimum_guessable_level"`
	ExcludedKeywords      []string `json:"excluded_keywords,omitempty" yaml:"excluded_keywords" msg:"excluded_keywords"`
	HistorySize           int      `json:"history_size,omitempty" yaml:"history_size" msg:"history_size"`
	HistoryDays           int      `json:"history_days,omitempty" yaml:"history_days" msg:"history_days"`
	ExpiryDays            int      `json:"expiry_days,omitempty" yaml:"expiry_days" msg:"expiry_days"`
}

func (c *PasswordPolicyConfiguration) IsPasswordHistoryEnabled() bool {
	return c.HistorySize > 0 || c.HistoryDays > 0
}

type AuthenticatorTOTPConfiguration struct {
	Maximum *int `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
}

type AuthenticatorOOBConfiguration struct {
	SMS   *AuthenticatorOOBSMSConfiguration   `json:"sms,omitempty" yaml:"sms" msg:"sms" default_zero_value:"true"`
	Email *AuthenticatorOOBEmailConfiguration `json:"email,omitempty" yaml:"email" msg:"email" default_zero_value:"true"`
}

type AuthenticatorOOBSMSConfiguration struct {
	Maximum *int                    `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
	Message SMSMessageConfiguration `json:"message,omitempty" yaml:"message" msg:"message" default_zero_value:"true"`
}

type AuthenticatorOOBEmailConfiguration struct {
	Maximum *int                      `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
	Message EmailMessageConfiguration `json:"message,omitempty" yaml:"message" msg:"message" default_zero_value:"true"`
}

type AuthenticatorBearerTokenConfiguration struct {
	ExpireInDays int `json:"expire_in_days,omitempty" yaml:"expire_in_days" msg:"expire_in_days"`
}

type AuthenticatorRecoveryCodeConfiguration struct {
	Count       int  `json:"count,omitempty" yaml:"count" msg:"count"`
	ListEnabled bool `json:"list_enabled,omitempty" yaml:"list_enabled" msg:"list_enabled"`
}
