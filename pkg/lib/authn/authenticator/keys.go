package authenticator

const (
	// AuthenticatorPropTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorPropTOTPDisplayName string = "https://authgear.com/claims/totp/display_name"
)

const (
	// AuthenticatorPropOOBOTPChannelType is a claim with string value for OOB OTP channel type.
	AuthenticatorPropOOBOTPChannelType string = "https://authgear.com/claims/oob_otp/channel_type"
	// AuthenticatorPropOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorPropOOBOTPEmail string = "https://authgear.com/claims/oob_otp/email"
	// AuthenticatorPropOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorPropOOBOTPPhone string = "https://authgear.com/claims/oob_otp/phone"
)

const (
	// AuthenticatorStateOOBOTPCode is a claim with string value for OOB OTP code secret of current interaction.
	// nolint:gosec
	AuthenticatorStateOOBOTPSecret string = "https://authgear.com/claims/oob_otp/secret"
)
