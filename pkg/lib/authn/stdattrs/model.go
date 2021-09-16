package stdattrs

type T map[string]interface{}

func (t T) ToClaims() map[string]interface{} {
	return map[string]interface{}(t)
}

const (
	// Sub is not used because we do not always use sub as the unique identifier for
	// an user from the identity provider.
	// Sub = "sub"
	Email             = "email"
	PhoneNumber       = "phone_number"
	PreferredUsername = "preferred_username"
	FamilyName        = "family_name"
	GivenName         = "given_name"
	Name              = "name"
	Nickname          = "nickname"
	Picture           = "picture"
	Profile           = "profile"
	Locale            = "locale"
)
