package principal

type Principal interface {
	PrincipalID() string
	PrincipalUserID() string
	ProviderID() string
	Attributes() Attributes
}

type Attributes interface{}

type Claims map[string]interface{}
