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

type AccessTokenDecoder interface {
	DecodeAccessToken(encodedToken string) (tok string, isHash bool, err error)
}

type Resolver struct {
	TrustProxy         config.TrustProxy
	Authorizations     AuthorizationStore
	AccessGrants       AccessGrantStore
	OfflineGrants      OfflineGrantStore
	AppSessions        AppSessionStore
	AccessTokenDecoder AccessTokenDecoder
	Sessions           ResolverSessionProvider
	SessionCookie      session.CookieDef
	Clock              clock.Clock
}

func (re *Resolver) Resolve(rw http.ResponseWriter, r *http.Request) (session.Session, error) {
	s, err := re.resolveHeader(r)
	if errors.Is(err, session.ErrInvalidSession) {
		s = nil
	} else if err != nil {
		return nil, err
	}
	if s != nil {
		return s, nil
	}

	s, err = re.resolveCookie(r)
	return s, err
}

func (re *Resolver) resolveHeader(r *http.Request) (session.Session, error) {
	token := parseAuthorizationHeader(r)
	if token == "" {
		// No bearer token in Authorization header. Simply proceed.
		return nil, nil
	}

	tok, isHash, err := re.AccessTokenDecoder.DecodeAccessToken(token)
	if err != nil {
		return nil, session.ErrInvalidSession
	}

	var tokenHash string
	if isHash {
		tokenHash = tok
	} else {
		tokenHash = HashToken(token)
	}

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
	event := access.NewEvent(re.Clock.NowUTC(), r, bool(re.TrustProxy))

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

func (re *Resolver) resolveCookie(r *http.Request) (session.Session, error) {
	cookie, err := r.Cookie(re.SessionCookie.Def.Name)
	if err != nil {
		// No session cookie. Simply proceed.
		return nil, nil
	}

	aSession, err := re.AppSessions.GetAppSession(HashToken(cookie.Value))
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	offlineGrant, err := re.OfflineGrants.GetOfflineGrant(aSession.OfflineGrantID)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	authz, err := re.Authorizations.GetByID(offlineGrant.AuthorizationID)
	if errors.Is(err, ErrAuthorizationNotFound) {
		// Authorization does not exists (e.g. revoked)
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	} else if !authz.IsAuthorized([]string{FullAccessScope}) {
		// App sessions must have full user access to be valid
		return nil, session.ErrInvalidSession
	}

	event := access.NewEvent(re.Clock.NowUTC(), r, bool(re.TrustProxy))
	offlineGrant.AccessInfo.LastAccess = event
	if err = re.OfflineGrants.UpdateOfflineGrant(offlineGrant); err != nil {
		return nil, err
	}

	return offlineGrant, nil
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
