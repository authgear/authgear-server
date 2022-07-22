package authenticator

const (
	// AuthenticatorClaimPasswordPasswordHash is a claim with []byte value.
	AuthenticatorClaimPasswordPasswordHash string = "https://authgear.com/claims/password/password_hash"
)

const (
	// AuthenticatorClaimTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorClaimTOTPDisplayName string = "https://authgear.com/claims/totp/display_name"
	// AuthenticatorClaimTOTPSecret is a claim with string value.
	AuthenticatorClaimTOTPSecret string = "https://authgear.com/claims/totp/secret"
)

const (
	// AuthenticatorClaimOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorClaimOOBOTPEmail string = "https://authgear.com/claims/oob_otp/email"
	// AuthenticatorClaimOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorClaimOOBOTPPhone string = "https://authgear.com/claims/oob_otp/phone"
)
