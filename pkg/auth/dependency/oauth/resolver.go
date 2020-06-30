package oauth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/clock"
)

type ResolverSessionProvider interface {
	Get(id string) (*session.IDPSession, error)
	Update(*session.IDPSession) error
}

type Resolver struct {
	ServerConfig   *config.ServerConfig
	Authorizations AuthorizationStore
	AccessGrants   AccessGrantStore
	OfflineGrants  OfflineGrantStore
	Sessions       ResolverSessionProvider
	Clock          clock.Clock
}

func (re *Resolver) Resolve(rw http.ResponseWriter, r *http.Request) (auth.AuthSession, error) {
	token := parseAuthorizationHeader(r)
	if token == "" {
		// No bearer token in Authorization header. Simply proceed.
		return nil, nil
	}

	token, err := DecodeAccessToken(token)
	if err != nil {
		return nil, auth.ErrInvalidSession
	}

	tokenHash := HashToken(token)
	grant, err := re.AccessGrants.GetAccessGrant(tokenHash)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, auth.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	_, err = re.Authorizations.GetByID(grant.AuthorizationID)
	if errors.Is(err, ErrAuthorizationNotFound) {
		// Authorization does not exists (e.g. revoked)
		return nil, auth.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	var authSession auth.AuthSession
	event := auth.NewAccessEvent(re.Clock.NowUTC(), r, re.ServerConfig.TrustProxy)

	switch grant.SessionKind {
	case GrantSessionKindSession:
		s, err := re.Sessions.Get(grant.SessionID)
		if errors.Is(err, session.ErrSessionNotFound) {
			return nil, auth.ErrInvalidSession
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
			return nil, auth.ErrInvalidSession
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
