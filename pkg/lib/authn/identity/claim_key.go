package identity

const (
	// IdentityClaimOAuthProviderType is a claim with a string value.
	IdentityClaimOAuthProviderType string = "https://authgear.com/claims/oauth/provider_type"
	// IdentityClaimOAuthProviderAlias is a claim with a string value.
	IdentityClaimOAuthProviderAlias string = "https://authgear.com/claims/oauth/provider_alias"
	// IdentityClaimOAuthSubjectID is a claim with a string value like `1098765432`.
	IdentityClaimOAuthSubjectID string = "https://authgear.com/claims/oauth/subject_id"
	// IdentityClaimOAuthData is a claim with a map value containing raw OAuth provider profile.
	IdentityClaimOAuthProfile string = "https://authgear.com/claims/oauth/profile"

	// IdentityClaimLoginIDType is a claim with a string value indicating the type of login ID.
	IdentityClaimLoginIDType string = "https://authgear.com/claims/login_id/type"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the key of login ID.
	IdentityClaimLoginIDKey string = "https://authgear.com/claims/login_id/key"
	// IdentityClaimLoginIDOriginalValue is a claim with a string value indicating the value of original login ID.
	IdentityClaimLoginIDOriginalValue string = "https://authgear.com/claims/login_id/original_value"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the value of login ID.
	IdentityClaimLoginIDValue string = "https://authgear.com/claims/login_id/value"

	// IdentityClaimAnonymousKeyID is a claim with a string value containing anonymous key ID.
	IdentityClaimAnonymousKeyID string = "https://authgear.com/claims/anonymous/key_id"

	// IdentityClaimBiometricKeyID is a claim with a string value containing biometric key ID.
	IdentityClaimBiometricKeyID string = "https://authgear.com/claims/biometric/key_id"
	// IdentityClaimBiometricDeviceInfo is a claim with a map value containing device info.
	IdentityClaimBiometricDeviceInfo string = "https://authgear.com/claims/biometric/device_info"
	// IdentityClaimBiometricFormattedDeviceInfo is a claim with a string value indicating formatted device info for display.
	IdentityClaimBiometricFormattedDeviceInfo string = "https://authgear.com/claims/biometric/formatted_device_info"

	// IdentityClaimPasskeyCredentialID is a claim with a string value.
	// nolint: gosec
	IdentityClaimPasskeyCredentialID string = "https://authgear.com/claims/passkey/credential_id"
	// nolint: gosec
	IdentityClaimPasskeyDisplayName string = "https://authgear.com/claims/passkey/display_name"

	// IdentityClaimSIWEAddress is a claim with a string value.
	IdentityClaimSIWEAddress string = "https://authgear.com/claims/siwe/address"
	// IdentityClaimSIWEChainID is a claim with an interger value.
	IdentityClaimSIWEChainID string = "https://authgear.com/claims/siwe/chain_id"

	// IdentityClaimLDAPServerName is a claim with a string value.
	IdentityClaimLDAPServerName string = "https://authgear.com/claims/ldap/server_name"
	// IdentityClaimLDAPUserIDAttributeName is a claim with a string value.
	IdentityClaimLDAPUserIDAttributeName string = "https://authgear.com/claims/ldap/user_id_attribute_name"
	// IdentityClaimLDAPUserIDAttributeValue is a claim with a string value.
	IdentityClaimLDAPUserIDAttributeValue string = "https://authgear.com/claims/ldap/user_id_attribute_value"
	// IdentityClaimLDAPRawUserIDAttributeValue is a claim with a string value.
	IdentityClaimLDAPRawUserIDAttributeValue string = "https://authgear.com/claims/ldap/raw_user_id_attribute_value"
	// IdentityClaimLDAPAttributes is a claim with a map value.
	IdentityClaimLDAPAttributes string = "https://authgear.com/claims/ldap/attributes"
	// IdentityClaimLDAPRawAttributes is a claim with a map value.
	IdentityClaimLDAPRawAttributes string = "https://authgear.com/claims/ldap/raw_attributes"
)
