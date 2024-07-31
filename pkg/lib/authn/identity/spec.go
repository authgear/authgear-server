package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Spec struct {
	Type model.IdentityType `json:"type"`

	LoginID   *LoginIDSpec   `json:"login_id,omitempty"`
	OAuth     *OAuthSpec     `json:"oauth,omitempty"`
	Anonymous *AnonymousSpec `json:"anonymous,omitempty"`
	Biometric *BiometricSpec `json:"biometric,omitempty"`
	Passkey   *PasskeySpec   `json:"passkey,omitempty"`
	SIWE      *SIWESpec      `json:"siwe,omitempty"`
	LDAP      *LDAPSpec      `json:"ldap,omitempty"`
}
