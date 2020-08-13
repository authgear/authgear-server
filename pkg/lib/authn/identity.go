package authn

type IdentityType string

const (
	IdentityTypeLoginID   IdentityType = "login_id"
	IdentityTypeOAuth     IdentityType = "oauth"
	IdentityTypeAnonymous IdentityType = "anonymous"
)

// PrimaryAuthenticatorTypes returns a list of authenticator types that can be used with t.
func (t IdentityType) PrimaryAuthenticatorTypes() []AuthenticatorType {
	switch t {
	case IdentityTypeLoginID:
		return []AuthenticatorType{
			AuthenticatorTypePassword,
			AuthenticatorTypeOOB,
		}
	case IdentityTypeOAuth:
		return nil
	case IdentityTypeAnonymous:
		return nil
	default:
		panic("unexpected identity type: " + t)
	}
}

// MatchingAuthenticatorTypes returns a ist of matching authenticator types.
func (t IdentityType) MatchingAuthenticatorTypes() []AuthenticatorType {
	switch t {
	case IdentityTypeLoginID:
		return []AuthenticatorType{
			AuthenticatorTypeOOB,
		}
	case IdentityTypeOAuth:
		return nil
	case IdentityTypeAnonymous:
		return nil
	default:
		panic("unexpected identity type: " + t)
	}
}
