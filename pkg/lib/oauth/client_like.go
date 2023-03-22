package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

type ClientLike struct {
	IsFirstParty        bool
	PIIAllowedInIDToken bool
	Scopes              []string
}

var ClientLikeNotFound = &ClientLike{
	IsFirstParty:        false,
	PIIAllowedInIDToken: false,
}

func SessionClientLike(s session.Session, c *config.OAuthConfig) *ClientLike {
	scopes := SessionScopes(s)
	switch s := s.(type) {
	case *idpsession.IDPSession:
		return &ClientLike{
			IsFirstParty:        true,
			PIIAllowedInIDToken: false,
			Scopes:              scopes,
		}
	case *OfflineGrant:
		client, ok := c.GetClient(s.ClientID)
		if !ok {
			return ClientLikeNotFound
		}
		return ClientClientLike(client, scopes)
	default:
		panic("oauth: unexpected session type")
	}
}

func ClientClientLike(client *config.OAuthClientConfig, scopes []string) *ClientLike {
	return &ClientLike{
		IsFirstParty:        client.IsFirstParty(),
		PIIAllowedInIDToken: client.PIIAllowedInIDToken(),
		Scopes:              scopes,
	}
}
