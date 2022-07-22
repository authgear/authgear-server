package identity

type ClaimKey string

const (
	// IdentityClaimOAuthProviderKeys is a claim with a map value like `{ "type": "azureadv2", "tenant": "test" }`.
	IdentityClaimOAuthProviderKeys ClaimKey = "https://authgear.com/claims/oauth/provider_keys"

	// IdentityClaimOAuthProviderType is a claim with a string value.
	IdentityClaimOAuthProviderType ClaimKey = "https://authgear.com/claims/oauth/provider_type"
	// IdentityClaimOAuthProviderAlias is a claim with a string value.
	IdentityClaimOAuthProviderAlias ClaimKey = "https://authgear.com/claims/oauth/provider_alias"
	// IdentityClaimOAuthSubjectID is a claim with a string value like `1098765432`.
	IdentityClaimOAuthSubjectID ClaimKey = "https://authgear.com/claims/oauth/subject_id"
	// IdentityClaimOAuthData is a claim with a map value containing raw OAuth provider profile.
	IdentityClaimOAuthProfile ClaimKey = "https://authgear.com/claims/oauth/profile"
	// IdentityClaimOAuthData is a claim with a map value containing mapped OIDC claims.
	IdentityClaimOAuthClaims ClaimKey = "https://authgear.com/claims/oauth/claims"

	// IdentityClaimLoginIDType is a claim with a string value indicating the type of login ID.
	IdentityClaimLoginIDType ClaimKey = "https://authgear.com/claims/login_id/type"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the key of login ID.
	IdentityClaimLoginIDKey ClaimKey = "https://authgear.com/claims/login_id/key"
	// IdentityClaimLoginIDOriginalValue is a claim with a string value indicating the value of original login ID.
	IdentityClaimLoginIDOriginalValue ClaimKey = "https://authgear.com/claims/login_id/original_value"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the value of login ID.
	IdentityClaimLoginIDValue ClaimKey = "https://authgear.com/claims/login_id/value"
	// IdentityClaimLoginIDUniqueKey is a claim with a string value containing the unique normalized login ID.
	IdentityClaimLoginIDUniqueKey ClaimKey = "https://authgear.com/claims/login_id/unique_key"

	// IdentityClaimAnonymousExistingUserID and IdentityClaimAnonymousExistingIdentityID are used for retrieving the anonymous identity if the identity doesn't have key.
	// IdentityClaimAnonymousExistingUserID is a claim with a string value containing the existing anonymous user id.
	IdentityClaimAnonymousExistingUserID ClaimKey = "https://authgear.com/claims/anonymous/existing_user_id"
	// IdentityClaimAnonymousExistingIdentityID is a claim with a string value containing the existing anonymous identity id.
	IdentityClaimAnonymousExistingIdentityID ClaimKey = "https://authgear.com/claims/anonymous/existing_identity_id"
	// IdentityClaimAnonymousKeyID is a claim with a string value containing anonymous key ID.
	IdentityClaimAnonymousKeyID ClaimKey = "https://authgear.com/claims/anonymous/key_id"
	// IdentityClaimAnonymousKey is a claim with a string value containing anonymous public key JWK.
	IdentityClaimAnonymousKey ClaimKey = "https://authgear.com/claims/anonymous/key"

	// IdentityClaimBiometricKeyID is a claim with a string value containing biometric key ID.
	IdentityClaimBiometricKeyID ClaimKey = "https://authgear.com/claims/biometric/key_id"
	// IdentityClaimBiometricKey is a claim with a string value containing biometric public key JWK.
	IdentityClaimBiometricKey ClaimKey = "https://authgear.com/claims/biometric/key"
	// IdentityClaimBiometricDeviceInfo is a claim with a map value containing device info.
	IdentityClaimBiometricDeviceInfo ClaimKey = "https://authgear.com/claims/biometric/device_info"
	// IdentityClaimBiometricFormattedDeviceInfo is a claim with a string value indicating formatted device info for display.
	IdentityClaimBiometricFormattedDeviceInfo ClaimKey = "https://authgear.com/claims/biometric/formatted_device_info"

	// IdentityClaimPasskeyCredentialID is a claim with a string value.
	// nolint: gosec
	IdentityClaimPasskeyCredentialID ClaimKey = "https://authgear.com/claims/passkey/credential_id"
	// IdentityClaimPasskeyCreationOptions ia a claim with a *CreationOption value.
	// nolint: gosec
	IdentityClaimPasskeyCreationOptions ClaimKey = "https://authgear.com/claims/passkey/creation_options"
	// IdentityClaimPasskeyAttestationResponse ia a claim with a []byte value.
	// nolint: gosec
	IdentityClaimPasskeyAttestationResponse ClaimKey = "https://authgear.com/claims/passkey/attestation_response"
	// IdentityClaimPasskeyAssertionResponse ia a claim with a []byte value.
	// nolint: gosec
	IdentityClaimPasskeyAssertionResponse ClaimKey = "https://authgear.com/claims/passkey/assertion_response"

	StandardClaimEmail             ClaimKey = "email"
	StandardClaimPhoneNumber       ClaimKey = "phone_number"
	StandardClaimPreferredUsername ClaimKey = "preferred_username"
)

func (k ClaimKey) IsPublic() bool {
	switch k {
	case IdentityClaimOAuthProviderType,
		IdentityClaimOAuthProviderAlias,
		IdentityClaimOAuthSubjectID,
		IdentityClaimOAuthProfile,
		IdentityClaimLoginIDType,
		IdentityClaimLoginIDKey,
		IdentityClaimLoginIDOriginalValue,
		IdentityClaimLoginIDValue,
		IdentityClaimAnonymousKeyID,
		IdentityClaimBiometricKeyID,
		IdentityClaimBiometricDeviceInfo,
		IdentityClaimBiometricFormattedDeviceInfo,
		IdentityClaimPasskeyCredentialID,
		StandardClaimEmail,
		StandardClaimPhoneNumber,
		StandardClaimPreferredUsername:
		return true
	default:
		return false
	}
}
