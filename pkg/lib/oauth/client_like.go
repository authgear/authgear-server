package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

type ClientLike struct {
	ClientParty config.ClientParty
	Scopes      []string
}

var ClientLikeNotFound = &ClientLike{
	ClientParty: config.ClientPartyThird,
}

func SessionClientLike(s session.Session, c *config.OAuthConfig) *ClientLike {
	scopes := SessionScopes(s)
	switch s := s.(type) {
	case *idpsession.IDPSession:
		return &ClientLike{
			ClientParty: config.ClientPartyFirst,
			Scopes:      scopes,
		}
	case *OfflineGrant:
		client, ok := c.GetClient(s.ClientID)
		if !ok {
			return ClientLikeNotFound
		}
		return &ClientLike{
			ClientParty: client.ClientParty(),
			Scopes:      scopes,
		}
	default:
		panic("oauth: unexpected session type")
	}
}
