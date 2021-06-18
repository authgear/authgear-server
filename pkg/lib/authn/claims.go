package authn

type ClaimName string

// ref: https://www.iana.org/assignments/jwt/jwt.xhtml
const (
	ClaimAMR               ClaimName = "amr"
	ClaimEmail             ClaimName = "email"
	ClaimPhoneNumber       ClaimName = "phone_number"
	ClaimPreferredUsername ClaimName = "preferred_username"
	ClaimKeyID             ClaimName = "https://authgear.com/claims/user/key_id"
	ClaimUserIsAnonymous   ClaimName = "https://authgear.com/claims/user/is_anonymous"
	ClaimUserIsVerified    ClaimName = "https://authgear.com/claims/user/is_verified"
)
