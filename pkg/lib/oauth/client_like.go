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

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

func SessionClientLike(s session.Session, clientResolver OAuthClientResolver) *ClientLike {
	scopes := SessionScopes(s)
	switch s := s.(type) {
	case *idpsession.IDPSession:
		return &ClientLike{
			IsFirstParty:        true,
			PIIAllowedInIDToken: false,
			Scopes:              scopes,
		}
	case *OfflineGrant:
		client := clientResolver.ResolveClient(s.ClientID)
		if client == nil {
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
