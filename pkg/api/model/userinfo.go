package model

import "time"

type UserInfoAuthenticator struct {
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Type      AuthenticatorType `json:"type"`
	Kind      AuthenticatorKind `json:"kind"`
}

type UserInfoIdentity struct {
	Type          IdentityType `json:"type"`
	ProviderAlias string       `json:"provider_alias,omitempty"`
}
