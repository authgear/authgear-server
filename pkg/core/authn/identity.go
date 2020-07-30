package authn

type IdentityType string

const (
	IdentityTypeLoginID   IdentityType = "login_id"
	IdentityTypeOAuth     IdentityType = "oauth"
	IdentityTypeAnonymous IdentityType = "anonymous"
)

// AssociatedAuthenticatorTypes returns a list of authenticator types that can be used with t.
func (t IdentityType) AssociatedAuthenticatorTypes() []AuthenticatorType {
	switch t {
	case IdentityTypeLoginID:
		return []AuthenticatorType{
			AuthenticatorTypePassword,
			AuthenticatorTypeTOTP,
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
