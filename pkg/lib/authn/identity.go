package authn

type IdentityType string

const (
	IdentityTypeLoginID   IdentityType = "login_id"
	IdentityTypeOAuth     IdentityType = "oauth"
	IdentityTypeAnonymous IdentityType = "anonymous"
	IdentityTypeBiometric IdentityType = "biometric"
)
