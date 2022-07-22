package authenticator

type ClaimKey string

const (
	// AuthenticatorClaimPasswordPasswordHash is a claim with []byte value.
	AuthenticatorClaimPasswordPasswordHash ClaimKey = "https://authgear.com/claims/password/password_hash"
	// AuthenticatorClaimPasswordPlainPassword is a claim with string value.
	AuthenticatorClaimPasswordPlainPassword ClaimKey = "https://authgear.com/claims/password/plain_password"
)

const (
	// AuthenticatorClaimTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorClaimTOTPDisplayName ClaimKey = "https://authgear.com/claims/totp/display_name"
	// AuthenticatorClaimTOTPSecret is a claim with string value.
	AuthenticatorClaimTOTPSecret ClaimKey = "https://authgear.com/claims/totp/secret"
)

const (
	// AuthenticatorClaimOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorClaimOOBOTPEmail ClaimKey = "https://authgear.com/claims/oob_otp/email"
	// AuthenticatorClaimOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorClaimOOBOTPPhone ClaimKey = "https://authgear.com/claims/oob_otp/phone"
)

func (k ClaimKey) IsPublic() bool {
	switch k {
	case AuthenticatorClaimTOTPDisplayName,
		AuthenticatorClaimOOBOTPEmail,
		AuthenticatorClaimOOBOTPPhone:
		return true
	default:
		return false
	}
}
