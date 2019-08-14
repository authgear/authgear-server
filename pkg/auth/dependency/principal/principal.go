package principal

type Principal interface {
	PrincipalID() string
	PrincipalUserID() string
	ProviderID() string
	Attributes() Attributes
	Claims() Claims
}

type Attributes map[string]interface{}

type Claims map[string]interface{}
