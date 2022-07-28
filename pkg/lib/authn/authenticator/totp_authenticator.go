package authenticator

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type TOTP struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Kind        string    `json:"kind"`
	IsDefault   bool      `json:"is_default"`
	Secret      string    `json:"secret"`
	DisplayName string    `json:"display_name"`
}

func (a *TOTP) ToInfo() *Info {
	return &Info{
		ID:        a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      model.AuthenticatorTypeTOTP,
		Kind:      Kind(a.Kind),
		IsDefault: a.IsDefault,

		TOTP: a,
	}
}
