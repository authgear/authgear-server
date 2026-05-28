package model

import "time"

type UserInfoAuthenticator struct {
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Type      AuthenticatorType `json:"type"`
	Kind      AuthenticatorKind `json:"kind"`
}

type UserInfoIdentity struct {
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Type          IdentityType   `json:"type"`
	LoginIDKey    string         `json:"login_id_key,omitempty"`
	LoginIDType   LoginIDKeyType `json:"login_id_type,omitempty"`
	ProviderAlias string         `json:"provider_alias,omitempty"`
}
