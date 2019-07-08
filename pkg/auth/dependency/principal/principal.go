package principal

type Principal interface {
	PrincipalID() string
	PrincipalUserID() string
	ProviderType() string
	Attributes() Attributes
}

type Attributes interface{}
