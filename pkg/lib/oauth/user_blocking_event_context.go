package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type UserBlockingEventContextUserService interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type UserBlockingEventContextIdentityService interface {
	ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error)
}

// UserBlockingEventContext holds the user-derived values needed to populate the
// oidc.id_token.pre_create and oidc.jwt.pre_create blocking event payloads.
//
// A single token request prepares both events for the same user, so the values
// are computed once and passed explicitly to both preparation calls. This is a
// value carried within one request, not a cache: it is not stored, not reused
// across requests, and not revalidated. It is only safe because both
// preparations happen inside one transaction, before any hook can run and
// mutate the user.
type UserBlockingEventContext struct {
	UserID     string
	UserModel  *model.User
	Identities []model.Identity
}

// GetUserModel and GetIdentities are nil-receiver safe so call sites do not
// branch on whether the context was computed.
func (c *UserBlockingEventContext) GetUserModel() *model.User {
	if c == nil {
		return nil
	}
	return c.UserModel
}

func (c *UserBlockingEventContext) GetIdentities() []model.Identity {
	if c == nil {
		return nil
	}
	return c.Identities
}

type UserBlockingEventContextProvider struct {
	Users      UserBlockingEventContextUserService
	Identities UserBlockingEventContextIdentityService
}

// Get reads the identity list and the user model. It issues 4 identity queries
// plus the 7 queries of user.Queries.GetMany.
func (p *UserBlockingEventContextProvider) Get(ctx context.Context, userID string) (*UserBlockingEventContext, error) {
	identities, err := p.Identities.ListIdentitiesThatHaveStandardAttributes(ctx, userID)
	if err != nil {
		return nil, err
	}

	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}

	u, err := p.Users.Get(ctx, userID, accesscontrol.RoleGreatest)
	if err != nil {
		return nil, err
	}

	return &UserBlockingEventContext{
		UserID:     userID,
		UserModel:  u,
		Identities: identityModels,
	}, nil
}
