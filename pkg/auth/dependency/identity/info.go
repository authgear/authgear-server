package identity

import (
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type Info struct {
	ID       string                 `json:"id"`
	Type     authn.IdentityType     `json:"type"`
	Claims   map[string]interface{} `json:"claims"`
	Identity interface{}            `json:"-"`
}

func (i *Info) ToSpec() Spec {
	return Spec{Type: i.Type, Claims: i.Claims}
}

func (i *Info) ToRef() Ref {
	return Ref{ID: i.ID, Type: i.Type}
}

func (i *Info) ToModel() model.Identity {
	claims := make(map[string]interface{})
	for key, value := range i.Claims {
		switch key {
		// It contains client_id, tenant or team_id, which should not
		// be exposed to clients.
		case IdentityClaimOAuthProviderKeys:
			continue

		// It contains OIDC standard claims, which is already exposed
		// as top-level claims.
		case IdentityClaimOAuthClaims:
			continue

		// It is a implementation details of login ID normalization,
		// so it should not be used by clients.
		case IdentityClaimLoginIDUniqueKey:
			continue

		// It is not useful to clients, since key ID should be
		// sufficient to identify a key.
		case IdentityClaimAnonymousKey:
			continue

		}
		claims[key] = value
	}

	return model.Identity{
		Type:   string(i.Type),
		Claims: claims,
	}
}
