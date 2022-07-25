package authenticator

type ClaimKey string

const (
	// AuthenticatorClaimPasswordPasswordHash is a claim with []byte value.
	// nolint: gosec
	AuthenticatorClaimPasswordPasswordHash ClaimKey = "https://authgear.com/claims/password/password_hash"
	// AuthenticatorClaimPasswordPlainPassword is a claim with string value.
	// nolint: gosec
	AuthenticatorClaimPasswordPlainPassword ClaimKey = "https://authgear.com/claims/password/plain_password"
)

const (
	// AuthenticatorClaimTOTPDisplayName is a claim with string value for TOTP display name.
	AuthenticatorClaimTOTPDisplayName ClaimKey = "https://authgear.com/claims/totp/display_name"
	// AuthenticatorClaimTOTPSecret is a claim with string value.
	// nolint: gosec
	AuthenticatorClaimTOTPSecret ClaimKey = "https://authgear.com/claims/totp/secret"
	// AuthenticatorClaimTOTPCode is a claim with string value.
	AuthenticatorClaimTOTPCode ClaimKey = "https://authgear.com/claims/totp/code"
)

const (
	// AuthenticatorClaimOOBOTPEmail is a claim with string value for OOB OTP email channel.
	AuthenticatorClaimOOBOTPEmail ClaimKey = "https://authgear.com/claims/oob_otp/email"
	// AuthenticatorClaimOOBOTPPhone is a claim with string value for OOB OTP phone channel.
	AuthenticatorClaimOOBOTPPhone ClaimKey = "https://authgear.com/claims/oob_otp/phone"
	// AuthenticatorClaimOOBOTPCode is a claim with string value.
	AuthenticatorClaimOOBOTPCode ClaimKey = "https://authgear.com/claims/oob_otp/code"
)

const (
	// AuthenticatorClaimPasskeyCredentialID is a claim with a string value.
	// nolint: gosec
	AuthenticatorClaimPasskeyCredentialID ClaimKey = "https://authgear.com/claims/passkey/credential_id"
	// AuthenticatorClaimPasskeyCreationOptions ia a claim with a *CreationOption value.
	// nolint: gosec
	AuthenticatorClaimPasskeyCreationOptions ClaimKey = "https://authgear.com/claims/passkey/creation_options"
	// AuthenticatorClaimPasskeyAttestationResponse ia a claim with a []byte value.
	// nolint: gosec
	AuthenticatorClaimPasskeyAttestationResponse ClaimKey = "https://authgear.com/claims/passkey/attestation_response"
	// AuthenticatorClaimPasskeyAssertionResponse ia a claim with a []byte value.
	// nolint: gosec
	AuthenticatorClaimPasskeyAssertionResponse ClaimKey = "https://authgear.com/claims/passkey/assertion_response"

	// AuthenticatorClaimPasskeySignCount is a claim with int64 value.
	// nolint: gosec
	AuthenticatorClaimPasskeySignCount ClaimKey = "https://authgear.com/claims/passkey/sign_count"
)

func (k ClaimKey) IsPublic() bool {
	switch k {
	case AuthenticatorClaimTOTPDisplayName,
		AuthenticatorClaimOOBOTPEmail,
		AuthenticatorClaimOOBOTPPhone,
		AuthenticatorClaimPasskeyCredentialID:
		return true
	default:
		return false
	}
}
