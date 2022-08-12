package identity

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type LoginID struct {
	ID              string                 `json:"id"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	UserID          string                 `json:"user_id"`
	LoginIDKey      string                 `json:"login_id_key"`
	LoginIDType     model.LoginIDKeyType   `json:"login_id_type"`
	LoginID         string                 `json:"login_id"`
	OriginalLoginID string                 `json:"original_login_id"`
	UniqueKey       string                 `json:"unique_key"`
	Claims          map[string]interface{} `json:"claims,omitempty"`
}

func (i *LoginID) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeLoginID,

		LoginID: i,
	}
}
