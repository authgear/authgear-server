package user

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type User struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
	Metadata    map[string]interface{}
}

func newUserModel(
	user *User,
	identities []*identity.Info,
) *model.User {
	isAnonymous := false
	for _, i := range identities {
		if i.Type == authn.IdentityTypeAnonymous {
			isAnonymous = true
			break
		}
	}

	return &model.User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
		IsAnonymous: isAnonymous,
		Metadata:    user.Metadata,
	}
}
