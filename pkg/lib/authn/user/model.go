package user

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type SortBy string

const (
	SortByDefault     SortBy = ""
	SortByCreatedAt   SortBy = "created_at"
	SortByLastLoginAt SortBy = "last_login_at"
)

type SortOption struct {
	SortBy        SortBy
	SortDirection model.SortDirection
}

func (o SortOption) Apply(builder db.SelectBuilder) db.SelectBuilder {
	sortBy := o.SortBy
	if sortBy == SortByDefault {
		sortBy = SortByCreatedAt
	}

	sortDirection := o.SortDirection
	if sortDirection == model.SortDirectionDefault {
		sortDirection = model.SortDirectionDesc
	}

	return builder.OrderBy(fmt.Sprintf("%s %s NULLS LAST", sortBy, sortDirection))
}

type User struct {
	ID                 string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	MostRecentLoginAt  *time.Time
	LessRecentLoginAt  *time.Time
	IsDisabled         bool
	DisableReason      *string
	StandardAttributes map[string]interface{}
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
	authenticators []*authenticator.Info,
	isVerified bool,
) *model.User {
	isAnonymous := false
	for _, i := range identities {
		if i.Type == authn.IdentityTypeAnonymous {
			isAnonymous = true
			break
		}
	}

	canReauthenticate := false
	for _, i := range authenticators {
		if i.Kind == authenticator.KindPrimary {
			canReauthenticate = true
		}
		if i.Kind == authenticator.KindSecondary {
			canReauthenticate = true
		}
	}

	return &model.User{
		Meta: model.Meta{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		LastLoginAt:       user.MostRecentLoginAt,
		IsAnonymous:       isAnonymous,
		IsVerified:        isVerified,
		IsDisabled:        user.IsDisabled,
		CanReauthenticate: canReauthenticate,
		DisableReason:     user.DisableReason,
	}
}
