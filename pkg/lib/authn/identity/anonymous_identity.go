package identity

import (
	"time"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Anonymous struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    string    `json:"user_id"`
	KeyID     string    `json:"key_id"`
	Key       []byte    `json:"key"`
}

func (i *Anonymous) ToJWK() (jwk.Key, error) {
	return jwk.ParseKey(i.Key)
}

func (i *Anonymous) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeAnonymous,

		Anonymous: i,
	}
}
