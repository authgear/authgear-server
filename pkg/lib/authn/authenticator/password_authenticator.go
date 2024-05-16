package authenticator

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Password struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Kind         string     `json:"kind"`
	IsDefault    bool       `json:"is_default"`
	PasswordHash []byte     `json:"password_hash,omitempty"`
	ExpireAfter  *time.Time `json:"expire_after,omitempty"`
}

func (a *Password) ToInfo() *Info {
	return &Info{
		ID:        a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      model.AuthenticatorTypePassword,
		Kind:      Kind(a.Kind),
		IsDefault: a.IsDefault,

		Password: a,
	}
}
