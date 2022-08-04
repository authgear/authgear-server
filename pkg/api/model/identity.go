package model

import (
	"fmt"
)

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

func (t IdentityType) PrimaryAuthenticatorTypes(loginIDKeyType LoginIDKeyType) []AuthenticatorType {
	switch t {
	case IdentityTypeLoginID:
		switch loginIDKeyType {
		case LoginIDKeyTypeUsername:
			return []AuthenticatorType{
				AuthenticatorTypePassword,
				AuthenticatorTypePasskey,
			}
		case LoginIDKeyTypeEmail:
			return []AuthenticatorType{
				AuthenticatorTypePassword,
				AuthenticatorTypePasskey,
				AuthenticatorTypeOOBEmail,
			}
		case LoginIDKeyTypePhone:
			return []AuthenticatorType{
				AuthenticatorTypePassword,
				AuthenticatorTypePasskey,
				AuthenticatorTypeOOBSMS,
			}
		default:
			panic(fmt.Sprintf("identity: unexpected login ID type: %s", loginIDKeyType))
		}
	case IdentityTypeOAuth:
		return nil
	case IdentityTypeAnonymous:
		return nil
	case IdentityTypeBiometric:
		return nil
	case IdentityTypePasskey:
		return []AuthenticatorType{
			AuthenticatorTypePasskey,
		}
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", t))
	}
}

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
