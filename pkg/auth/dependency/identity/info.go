package identity

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
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
		// Hide IdentityClaimOAuthProviderKeys because
		// It may contain client_id, tenant or team_id.
		if key == IdentityClaimOAuthProviderKeys {
			continue
		}
		claims[key] = value
	}

	return model.Identity{
		Type:   string(i.Type),
		Claims: claims,
	}
}
