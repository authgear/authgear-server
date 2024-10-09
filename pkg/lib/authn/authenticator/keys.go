package authenticator

const (
	// AuthenticatorClaimTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorClaimTOTPDisplayName string = "https://authgear.com/claims/totp/display_name"
)

const (
	// AuthenticatorClaimOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorClaimOOBOTPEmail string = "https://authgear.com/claims/oob_otp/email"
	// AuthenticatorClaimOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorClaimOOBOTPPhone string = "https://authgear.com/claims/oob_otp/phone"
)

const (
	// AuthenticatorClaimPasskeyCredentialID is a claim with a string value.
	// nolint: gosec
	AuthenticatorClaimPasskeyCredentialID string = "https://authgear.com/claims/passkey/credential_id"
)

const (
	AuthenticatorClaimFaceRecognition string = "https://authgear.com/claims/face_recognition"
)
