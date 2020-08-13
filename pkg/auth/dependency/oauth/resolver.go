package oauth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type ResolverSessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
	Update(*idpsession.IDPSession) error
}

type Resolver struct {
	ServerConfig   *config.ServerConfig
	Authorizations AuthorizationStore
	AccessGrants   AccessGrantStore
	OfflineGrants  OfflineGrantStore
	Sessions       ResolverSessionProvider
	Clock          clock.Clock
}

func (re *Resolver) Resolve(rw http.ResponseWriter, r *http.Request) (session.Session, error) {
	token := parseAuthorizationHeader(r)
	if token == "" {
		// No bearer token in Authorization header. Simply proceed.
		return nil, nil
	}

	token, err := DecodeAccessToken(token)
	if err != nil {
		return nil, session.ErrInvalidSession
	}

	tokenHash := HashToken(token)
	grant, err := re.AccessGrants.GetAccessGrant(tokenHash)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	_, err = re.Authorizations.GetByID(grant.AuthorizationID)
	if errors.Is(err, ErrAuthorizationNotFound) {
		// Authorization does not exists (e.g. revoked)
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	var authSession session.Session
	event := access.NewEvent(re.Clock.NowUTC(), r, re.ServerConfig.TrustProxy)

	switch grant.SessionKind {
	case GrantSessionKindSession:
		s, err := re.Sessions.Get(grant.SessionID)
		if errors.Is(err, idpsession.ErrSessionNotFound) {
			return nil, session.ErrInvalidSession
		} else if err != nil {
			return nil, err
		}
		s.AccessInfo.LastAccess = event
		if err = re.Sessions.Update(s); err != nil {
			return nil, err
		}

		authSession = s

	case GrantSessionKindOffline:
		g, err := re.OfflineGrants.GetOfflineGrant(grant.SessionID)
		if errors.Is(err, ErrGrantNotFound) {
			return nil, session.ErrInvalidSession
		} else if err != nil {
			return nil, err
		}
		g.AccessInfo.LastAccess = event
		if err = re.OfflineGrants.UpdateOfflineGrant(g); err != nil {
			return nil, err
		}

		authSession = g

	default:
		panic("oauth: resolving unknown grant session kind")
	}

	return authSession, nil
}

func parseAuthorizationHeader(r *http.Request) (token string) {
	authorization := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(authorization) != 2 {
		return
	}

	scheme := authorization[0]
	if strings.ToLower(scheme) != "bearer" {
		return
	}

	return authorization[1]
}
