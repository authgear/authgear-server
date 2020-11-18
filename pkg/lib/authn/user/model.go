package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type User struct {
	ID            string
	Labels        map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	IsDisabled    bool
	DisableReason *string
}

func (u *User) GetMeta() model.Meta {
	return model.Meta{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (u *User) CheckStatus() error {
	if u.IsDisabled {
		return NewErrDisabledUser(u.DisableReason)
	}
	return nil
}

func newUserModel(
	user *User,
	identities []*identity.Info,
	isVerified bool,
) *model.User {
	isAnonymous := false
	for _, i := range identities {
		if i.Type == authn.IdentityTypeAnonymous {
			isAnonymous = true
			break
		}
	}

	return &model.User{
		Meta: model.Meta{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		LastLoginAt:   user.LastLoginAt,
		IsAnonymous:   isAnonymous,
		IsVerified:    isVerified,
		IsDisabled:    user.IsDisabled,
		DisableReason: user.DisableReason,
	}
}
