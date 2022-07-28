package model

type LoginIDKeyType string

const (
	LoginIDKeyTypeEmail    LoginIDKeyType = "email"
	LoginIDKeyTypePhone    LoginIDKeyType = "phone"
	LoginIDKeyTypeUsername LoginIDKeyType = "username"
)

var LoginIDKeyTypes = []LoginIDKeyType{
	LoginIDKeyTypeEmail,
	LoginIDKeyTypePhone,
	LoginIDKeyTypeUsername,
}

type IdentityType string

const (
	IdentityTypeLoginID   IdentityType = "login_id"
	IdentityTypeOAuth     IdentityType = "oauth"
	IdentityTypeAnonymous IdentityType = "anonymous"
	IdentityTypeBiometric IdentityType = "biometric"
	IdentityTypePasskey   IdentityType = "passkey"
)

type Identity struct {
	Meta
	Type   string                 `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

type IdentityRef struct {
	Meta
	UserID string
	Type   IdentityType
}

func (r *IdentityRef) ToRef() *IdentityRef { return r }
