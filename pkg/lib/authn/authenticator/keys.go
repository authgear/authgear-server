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
	// AuthenticatorClaimPasskeyAAGUID is a claim with a string value like `fbfc3007-154e-4ecc-8c0b-6e020557d7bd`.
	// nolint: gosec
	AuthenticatorClaimPasskeyAAGUID string = "https://authgear.com/claims/passkey/aaguid"
	// AuthenticatorClaimPasskeyProviderName is a claim with a string value like `Apple Passwords`.
	// It is absent when the AAGUID is unknown.
	// nolint: gosec
	AuthenticatorClaimPasskeyProviderName string = "https://authgear.com/claims/passkey/provider_name"
)
