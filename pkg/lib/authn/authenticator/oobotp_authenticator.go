package authenticator

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type OOBOTP struct {
	ID                   string                  `json:"id"`
	UserID               string                  `json:"user_id"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
	Kind                 string                  `json:"kind"`
	IsDefault            bool                    `json:"is_default"`
	OOBAuthenticatorType model.AuthenticatorType `json:"oob_authenticator_type"`
	Phone                string                  `json:"phone,omitempty"`
	Email                string                  `json:"email,omitempty"`
}

func (a *OOBOTP) ToInfo() *Info {
	return &Info{
		ID:        a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      a.OOBAuthenticatorType,
		Kind:      Kind(a.Kind),
		IsDefault: a.IsDefault,

		OOBOTP: a,
	}
}
