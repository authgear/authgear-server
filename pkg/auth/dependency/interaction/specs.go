package interaction

type IdentitySpec struct {
	ID     string                 `json:"-"`
	Type   IdentityType           `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

type AuthenticatorSpec struct {
	ID    string                 `json:"-"`
	Type  AuthenticatorType      `json:"type"`
	Props map[string]interface{} `json:"props"`
}
