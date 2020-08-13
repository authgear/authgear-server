package authn

type ClaimName string

// ref: https://www.iana.org/assignments/jwt/jwt.xhtml
const (
	ClaimACR               ClaimName = "acr"
	ClaimAMR               ClaimName = "amr"
	ClaimEmail             ClaimName = "email"
	ClaimPhoneNumber       ClaimName = "phone_number"
	ClaimPreferredUsername ClaimName = "preferred_username"
	ClaimKeyID             ClaimName = "https://authgear.com/user/key_id"
	ClaimUserIsAnonymous   ClaimName = "https://authgear.com/user/is_anonymous"
	ClaimUserIsVerified    ClaimName = "https://authgear.com/user/is_verified"
)
