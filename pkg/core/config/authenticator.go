package config

//go:generate msgp -tests=false
type AuthenticatorConfiguration struct {
	TOTP         *AuthenticatorTOTPConfiguration         `json:"totp,omitempty" yaml:"totp" msg:"totp" default_zero_value:"true"`
	OOB          *AuthenticatorOOBConfiguration          `json:"oob_otp,omitempty" yaml:"oob_otp" msg:"oob_otp" default_zero_value:"true"`
	BearerToken  *AuthenticatorBearerTokenConfiguration  `json:"bearer_token,omitempty" yaml:"bearer_token" msg:"bearer_token" default_zero_value:"true"`
	RecoveryCode *AuthenticatorRecoveryCodeConfiguration `json:"recovery_code,omitempty" yaml:"recovery_code" msg:"recovery_code" default_zero_value:"true"`
}

type AuthenticatorTOTPConfiguration struct {
	Maximum *int `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
}

type AuthenticatorOOBConfiguration struct {
	SMS     *AuthenticatorOOBSMSConfiguration   `json:"sms,omitempty" yaml:"sms" msg:"sms" default_zero_value:"true"`
	Email   *AuthenticatorOOBEmailConfiguration `json:"email,omitempty" yaml:"email" msg:"email" default_zero_value:"true"`
	Sender  string                              `json:"sender,omitempty" yaml:"sender" msg:"sender"`
	Subject string                              `json:"subject,omitempty" yaml:"subject" msg:"subject"`
	ReplyTo string                              `json:"reply_to,omitempty" yaml:"reply_to" msg:"reply_to"`
}

type AuthenticatorOOBSMSConfiguration struct {
	Maximum *int `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
}

type AuthenticatorOOBEmailConfiguration struct {
	Maximum *int `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
}

type AuthenticatorBearerTokenConfiguration struct {
	ExpireInDays int `json:"expire_in_days,omitempty" yaml:"expire_in_days" msg:"expire_in_days"`
}

type AuthenticatorRecoveryCodeConfiguration struct {
	Count       int  `json:"count,omitempty" yaml:"count" msg:"count"`
	ListEnabled bool `json:"list_enabled,omitempty" yaml:"list_enabled" msg:"list_enabled"`
}
