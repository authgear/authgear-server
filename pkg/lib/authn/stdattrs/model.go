package stdattrs

type T map[string]interface{}

func (t T) ToClaims() map[string]interface{} {
	return map[string]interface{}(t)
}

func (t T) WithNameCopiedToGivenName() T {
	out := make(T)
	for k, v := range t {
		out[k] = v
	}

	if name, ok := t[Name].(string); ok && name != "" {
		if _, ok := t[GivenName].(string); !ok {
			out[GivenName] = name
		}
	}
	return out
}

// NonIdentityAware returns a copy of t with identity-aware attributes removed.
func (t T) NonIdentityAware() T {
	out := make(T)
	for k1, val := range t {
		for _, k2 := range NonIdentityAwareKeys {
			if k1 == k2 {
				out[k1] = val
			}
		}
	}
	return out
}

// MergedWith returns a T with that merged into t.
func (t T) MergedWith(that T) T {
	out := make(T)
	for k, v := range t {
		out[k] = v
	}

	for k, v := range that {
		out[k] = v
	}

	return out
}

const (
	// Sub is not used because we do not always use sub as the unique identifier for
	// an user from the identity provider.
	// Sub = "sub"
	Email               = "email"
	EmailVerified       = "email_verified"
	PhoneNumber         = "phone_number"
	PhoneNumberVerified = "phone_number_verified"
	PreferredUsername   = "preferred_username"
	FamilyName          = "family_name"
	GivenName           = "given_name"
	MiddleName          = "middle_name"
	Name                = "name"
	Nickname            = "nickname"
	Picture             = "picture"
	Profile             = "profile"
	Website             = "website"
	Gender              = "gender"
	Birthdate           = "birthdate"
	Zoneinfo            = "zoneinfo"
	Locale              = "locale"
	Address             = "address"
	Formatted           = "formatted"
	StreetAddress       = "street_address"
	Locality            = "locality"
	Region              = "region"
	PostalCode          = "postal_code"
	Country             = "country"
)

var NonIdentityAwareKeys []string = []string{
	FamilyName,
	GivenName,
	MiddleName,
	Name,
	Nickname,
	Picture,
	Profile,
	Website,
	Gender,
	Birthdate,
	Zoneinfo,
	Locale,
	Address,
}
