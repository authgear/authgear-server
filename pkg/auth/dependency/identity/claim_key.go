package identity

const (
	// This claim is intended for internal use only.
	IdentityClaimOAuthProviderKeys string = "https://authgear.com/claims/oauth/provider_keys"
	// This claim is intended for internal use only.
	IdentityClaimOAuthAction string = "https://authgear.com/claims/oauth/action"
	// This claim is intended for internal use only.
	IdentityClaimOAuthNonce string = "https://authgear.com/claims/oauth/nonce"
	// This claim is intended for internal use only.
	IdentityClaimOAuthUserID string = "https://authgear.com/claims/oauth/user_id"
	// This claim is intended for internal use only.
	IdentityClaimOAuthGeneratedProviderRedirectURI string = "https://authgear.com/claims/oauth/generated_provider_redirect_uri"

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

	// IdentityClaimLoginIDValue is a claim with a string value indicating the key of login ID.
	IdentityClaimLoginIDKey string = "https://authgear.com/claims/login_id/key"
	// IdentityClaimLoginIDOriginalValue is a claim with a string value indicating the value of original login ID.
	IdentityClaimLoginIDOriginalValue string = "https://authgear.com/claims/login_id/original_value"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the value of login ID.
	IdentityClaimLoginIDValue string = "https://authgear.com/claims/login_id/value"
	// IdentityClaimLoginIDUniqueKey is a claim with a string value containing the unique normalized login ID.
	IdentityClaimLoginIDUniqueKey string = "https://authgear.com/claims/login_id/unique_key"

	// IdentityClaimAnonymousKeyID is a claim with a string value containing anonymous key ID.
	IdentityClaimAnonymousKeyID string = "https://authgear.com/claims/anonymous/key_id"
	// IdentityClaimAnonymousKey is a claim with a string value containing anonymous public key JWK.
	IdentityClaimAnonymousKey string = "https://authgear.com/claims/anonymous/key"
)
