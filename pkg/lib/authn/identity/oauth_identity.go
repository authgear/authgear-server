package identity

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OAuth struct {
	ID                string                 `json:"id"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	UserID            string                 `json:"user_id"`
	ProviderID        config.ProviderID      `json:"provider_id"`
	ProviderSubjectID string                 `json:"provider_subject_id"`
	UserProfile       map[string]interface{} `json:"user_profile,omitempty"`
	Claims            map[string]interface{} `json:"claims,omitempty"`
}

func (i *OAuth) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeOAuth,

		OAuth: i,
	}
}
