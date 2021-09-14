package stdattrs

type T map[string]interface{}

func (t T) Sub() string {
	sub, _ := t[Sub].(string)
	return sub
}

func (t T) ToClaims() map[string]interface{} {
	return map[string]interface{}(t)
}

const (
	Sub               = "sub"
	Email             = "email"
	PhoneNumber       = "phone_number"
	PreferredUsername = "preferred_username"
	FamilyName        = "family_name"
	GivenName         = "given_name"
	Name              = "name"
	Picture           = "picture"
	Profile           = "profile"
	Locale            = "locale"
)
