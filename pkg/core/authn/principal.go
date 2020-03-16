package authn

type PrincipalType string

const (
	PrincipalTypePassword PrincipalType = "password"
	PrincipalTypeOAuth    PrincipalType = "oauth"
)
