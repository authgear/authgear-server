package identity

const (
	// IdentityClaimOAuthProvider is a claim with a map value like `{ "type": "azureadv2", "tenant": "test" }`.
	IdentityClaimOAuthProvider string = "https://auth.skygear.io/claims/oauth/provider"
	// IdentityClaimOAuthSubjectID is a claim with a string value like `1098765432`.
	IdentityClaimOAuthSubjectID string = "https://auth.skygear.io/claims/oauth/subject_id"
	// IdentityClaimOAuthData is a claim with a map value containing raw OAuth provider profile.
	IdentityClaimOAuthProfile string = "https://auth.skygear.io/claims/oauth/profile"
	// IdentityClaimOAuthData is a claim with a map value containing mapped OIDC claims.
	IdentityClaimOAuthClaims string = "https://auth.skygear.io/claims/oauth/claims"

	// IdentityClaimLoginIDValue is a claim with a string value indicating the key of login ID.
	IdentityClaimLoginIDKey string = "https://auth.skygear.io/claims/login_id/key"
	// IdentityClaimLoginIDOriginalValue is a claim with a string value indicating the value of original login ID.
	IdentityClaimLoginIDOriginalValue string = "https://auth.skygear.io/claims/login_id/original_value"
	// IdentityClaimLoginIDValue is a claim with a string value indicating the value of login ID.
	IdentityClaimLoginIDValue string = "https://auth.skygear.io/claims/login_id/value"
	// IdentityClaimLoginIDUniqueKey is a claim with a string value containing the unique normalized login ID.
	IdentityClaimLoginIDUniqueKey string = "https://auth.skygear.io/claims/login_id/unique_key"

	// IdentityClaimAnonymousKeyID is a claim with a string value containing anonymous key ID.
	IdentityClaimAnonymousKeyID string = "https://auth.skygear.io/claims/anonymous/key_id"
	// IdentityClaimAnonymousKey is a claim with a string value containing anonymous public key JWK.
	IdentityClaimAnonymousKey string = "https://auth.skygear.io/claims/anonymous/key"
)
