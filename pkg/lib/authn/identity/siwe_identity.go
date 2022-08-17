package identity

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type SIWE struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    string    `json:"user_id"`
	ChainID   int       `json:"chain_id"`
	Address   string    `json:"address"`

	Data model.SIWEVerifiedData `json:"data"`
}

func (i *SIWE) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeSIWE,

		SIWE: i,
	}
}
