package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
)

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
}

type IdentityAdaptor struct {
	LoginID LoginIDIdentityProvider
	OAuth   OAuthIdentityProvider
}

func (a *IdentityAdaptor) Get(userID string, typ interaction.IdentityType, id string) (*interaction.IdentityInfo, error) {
	switch typ {
	case interaction.IdentityTypeLoginID:
		l, err := a.LoginID.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(l), nil

	case interaction.IdentityTypeOAuth:
		o, err := a.OAuth.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return oauthToIdentityInfo(o), nil
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}

func loginIDToIdentityInfo(l *loginid.Identity) *interaction.IdentityInfo {
	claims := map[string]interface{}{
		interaction.IdentityClaimLoginIDUniqueKey: l.UniqueKey,
	}
	for k, v := range l.Claims {
		claims[k] = v
	}

	return &interaction.IdentityInfo{
		Type:     interaction.IdentityTypeLoginID,
		ID:       l.ID,
		Claims:   claims,
		Identity: l,
	}
}
