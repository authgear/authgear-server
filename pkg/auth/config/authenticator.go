package config

type AuthenticatorConfig struct {
	Password     *AuthenticatorPasswordConfig     `json:"password,omitempty"`
	TOTP         *AuthenticatorTOTPConfig         `json:"totp,omitempty"`
	OOB          *AuthenticatorOOBConfig          `json:"oob_otp,omitempty"`
	BearerToken  *AuthenticatorBearerTokenConfig  `json:"bearer_token,omitempty"`
	RecoveryCode *AuthenticatorRecoveryCodeConfig `json:"recovery_code,omitempty"`
}

type AuthenticatorPasswordConfig struct {
	Policy *PasswordPolicyConfig `json:"policy,omitempty"`
}

type PasswordPolicyConfig struct {
	MinLength             int          `json:"min_length,omitempty"`
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

type AuthenticatorTOTPConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

type AuthenticatorOOBConfig struct {
	SMS   *AuthenticatorOOBSMSConfig   `json:"sms,omitempty"`
	Email *AuthenticatorOOBEmailConfig `json:"email,omitempty"`
}

type AuthenticatorOOBSMSConfig struct {
	Maximum *int             `json:"maximum,omitempty"`
	Message SMSMessageConfig `json:"message,omitempty"`
}

type AuthenticatorOOBEmailConfig struct {
	Maximum *int               `json:"maximum,omitempty"`
	Message EmailMessageConfig `json:"message,omitempty"`
}

type AuthenticatorBearerTokenConfig struct {
	ExpireIn DurationDays `json:"expire_in_days,omitempty"`
}

type AuthenticatorRecoveryCodeConfig struct {
	Count       int  `json:"count,omitempty"`
	ListEnabled bool `json:"list_enabled,omitempty"`
}
