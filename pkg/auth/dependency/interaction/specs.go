package interaction

type IdentitySpec struct {
	Type   IdentityType           `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

type AuthenticatorSpec struct {
	Type  AuthenticatorType      `json:"type"`
	Props map[string]interface{} `json:"props"`
}
