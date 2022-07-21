package identity

const (
	// This claim is intended for internal use only.
	IdentityClaimOAuthProviderKeys string = "https://authgear.com/claims/oauth/provider_keys"

	// IdentityClaimOAuthProviderType is a claim with a string value.
	// This claim is intended for external use only.
	IdentityClaimOAuthProviderType string = "https://authgear.com/claims/oauth/provider_type"
	// IdentityClaimOAuthProviderAlias is a claim with a string value.
	// This claim is intended for external use only.
	IdentityClaimOAuthProviderAlias string = "https://authgear.com/claims/oauth/provider_alias"
	// IdentityClaimOAuthProviderKeys is a claim with a map value like `{ "type": "azureadv2", "tenant": "test" }`.
	// IdentityClaimOAuthSubjectID is a claim with a string value like `1098765432`.
	IdentityClaimOAuthSubjectID string = "https://authgear.com/claims/oauth/subject_id"
	// IdentityClaimOAuthData is a claim with a map value containing raw OAuth provider profile.
	IdentityClaimOAuthProfile string = "https://authgear.com/claims/oauth/profile"
	// IdentityClaimOAuthData is a claim with a map value containing mapped OIDC claims.
	IdentityClaimOAuthClaims string = "https://authgear.com/claims/oauth/claims"

	// IdentityClaimLoginIDType is a claim with a string value indicating the type of login ID.
	IdentityClaimLoginIDType string = "https://authgear.com/claims/login_id/type"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the key of login ID.
	IdentityClaimLoginIDKey string = "https://authgear.com/claims/login_id/key"
	// IdentityClaimLoginIDOriginalValue is a claim with a string value indicating the value of original login ID.
	IdentityClaimLoginIDOriginalValue string = "https://authgear.com/claims/login_id/original_value"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the value of login ID.
	IdentityClaimLoginIDValue string = "https://authgear.com/claims/login_id/value"
	// IdentityClaimLoginIDUniqueKey is a claim with a string value containing the unique normalized login ID.
	IdentityClaimLoginIDUniqueKey string = "https://authgear.com/claims/login_id/unique_key"

	// IdentityClaimAnonymousExistingUserID and IdentityClaimAnonymousExistingIdentityID are used for retrieving the anonymous identity if the identity doesn't have key.
	// IdentityClaimAnonymousExistingUserID is a claim with a string value containing the existing anonymous user id.
	// This claim is intended for internal use.
	IdentityClaimAnonymousExistingUserID string = "https://authgear.com/claims/anonymous/existing_user_id"
	// IdentityClaimAnonymousExistingIdentityID is a claim with a string value containing the existing anonymous identity id.
	// This claim is intended for internal use.
	IdentityClaimAnonymousExistingIdentityID string = "https://authgear.com/claims/anonymous/existing_identity_id"
	// IdentityClaimAnonymousKeyID is a claim with a string value containing anonymous key ID.
	IdentityClaimAnonymousKeyID string = "https://authgear.com/claims/anonymous/key_id"
	// IdentityClaimAnonymousKey is a claim with a string value containing anonymous public key JWK.
	IdentityClaimAnonymousKey string = "https://authgear.com/claims/anonymous/key"

	// IdentityClaimBiometricKeyID is a claim with a string value containing biometric key ID.
	IdentityClaimBiometricKeyID string = "https://authgear.com/claims/biometric/key_id"
	// IdentityClaimBiometricKey is a claim with a string value containing biometric public key JWK.
	IdentityClaimBiometricKey string = "https://authgear.com/claims/biometric/key"
	// IdentityClaimBiometricDeviceInfo is a claim with a map value containing device info.
	IdentityClaimBiometricDeviceInfo string = "https://authgear.com/claims/biometric/device_info"
	// IdentityClaimBiometricFormattedDeviceInfo is a claim with a string value indicating formatted device info for display.
	IdentityClaimBiometricFormattedDeviceInfo string = "https://authgear.com/claims/biometric/formatted_device_info"

	// IdentityClaimPasskeyCredentialID is a claim with a string value.
	IdentityClaimPasskeyCredentialID string = "https://authgear.com/claims/passkey/credential_id"
	// IdentityClaimPasskeyCreationOptions ia a claim with a *CreationOption value.
	IdentityClaimPasskeyCreationOptions string = "https://authgear.com/claims/passkey/creation_options"
	// IdentityClaimPasskeyAttestationResponse ia a claim with a []byte value.
	IdentityClaimPasskeyAttestationResponse string = "https://authgear.com/claims/passkey/attestation_response"
)

const (
	StandardClaimEmail             string = "email"
	StandardClaimPhoneNumber       string = "phone_number"
	StandardClaimPreferredUsername string = "preferred_username"
)
