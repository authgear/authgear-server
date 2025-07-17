package oauth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type ResolverSessionProvider interface {
	AccessWithID(ctx context.Context, id string, accessEvent access.Event) (*idpsession.IDPSession, error)
}

type AccessTokenDecoder interface {
	DecodeAccessToken(encodedToken string) (tok string, isHash bool, err error)
}

type ResolverCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
}

type ResolverOfflineGrantService interface {
	AccessOfflineGrant(ctx context.Context, grantID string, refreshTokenHash string, accessEvent *access.Event, expireAt time.Time) (*OfflineGrant, error)
	GetOfflineGrant(ctx context.Context, id string) (*OfflineGrant, error)
}

type Resolver struct {
	RemoteIP            httputil.RemoteIP
	UserAgentString     httputil.UserAgentString
	OAuthConfig         *config.OAuthConfig
	Authorizations      AuthorizationStore
	AccessGrants        AccessGrantStore
	AppSessions         AppSessionStore
	AccessTokenDecoder  AccessTokenDecoder
	Sessions            ResolverSessionProvider
	Cookies             ResolverCookieManager
	Clock               clock.Clock
	OfflineGrantService ResolverOfflineGrantService
}

func (re *Resolver) Resolve(ctx context.Context, rw http.ResponseWriter, r *http.Request) (session.ResolvedSession, error) {
	// The resolve function has the following outcomes:
	// - (nil, nil) which means no session was found and no error.
	// - (nil, err) in which err is ErrInvalidSession, which means the session was found but was invalid.
	//              in which err is something else, which means some unexpected error occurred.
	// - (s, nil)  which means a session was resolved successfully.
	//
	// Here we want to try the next resolve function iff the outcome is (nil, nil).
	funcs := []func(context.Context, *http.Request) (session.ResolvedSession, error){
		re.resolveHeader,
		re.resolveAppSessionCookie,
		re.resolveAccessTokenCookie,
	}

	for _, f := range funcs {
		s, err := f(ctx, r)
		if err != nil {
			if errors.Is(err, session.ErrInvalidSession) {
				continue
			} else {
				return nil, err
			}
		}
		if s != nil {
			return s, nil
		}
	}

	return nil, nil
}

func (re *Resolver) resolveAccessToken(ctx context.Context, token string) (session.ResolvedSession, error) {
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

	grant, err := re.AccessGrants.GetAccessGrant(ctx, tokenHash)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	_, err = re.Authorizations.GetByID(ctx, grant.AuthorizationID)
	if errors.Is(err, ErrAuthorizationNotFound) {
		// Authorization does not exists (e.g. revoked)
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	var authSession session.ResolvedSession
	event := access.NewEvent(re.Clock.NowUTC(), re.RemoteIP, re.UserAgentString)

	switch grant.SessionKind {
	case GrantSessionKindSession:
		s, err := re.Sessions.AccessWithID(ctx, grant.SessionID, event)
		if errors.Is(err, idpsession.ErrSessionNotFound) {
			return nil, session.ErrInvalidSession
		} else if err != nil {
			return nil, err
		}
		authSession = s

	case GrantSessionKindOffline:
		g, err := re.OfflineGrantService.GetOfflineGrant(ctx, grant.SessionID)
		if errors.Is(err, ErrGrantNotFound) {
			return nil, session.ErrInvalidSession
		} else if err != nil {
			return nil, err
		}

		g, err = re.accessOfflineGrant(ctx, g, grant.RefreshTokenHash, event)
		if err != nil {
			return nil, err
		}

		as, ok := g.ToSession(grant.RefreshTokenHash)
		if !ok {
			return nil, session.ErrInvalidSession
		}
		authSession = as
	default:
		panic("oauth: resolving unknown grant session kind")
	}

	return authSession, nil
}

func (re *Resolver) resolveHeader(ctx context.Context, r *http.Request) (session.ResolvedSession, error) {
	token := parseAuthorizationHeader(r)
	if token == "" {
		// No bearer token in Authorization header. Simply proceed.
		return nil, nil
	}

	return re.resolveAccessToken(ctx, token)
}

func (re *Resolver) resolveAccessTokenCookie(ctx context.Context, r *http.Request) (session.ResolvedSession, error) {
	cookie, err := re.Cookies.GetCookie(r, session.AppAccessTokenCookieDef)
	if err != nil {
		// No access token cookie. Simply proceed.
		return nil, nil
	}

	return re.resolveAccessToken(ctx, cookie.Value)
}

func (re *Resolver) resolveAppSessionCookie(ctx context.Context, r *http.Request) (session.ResolvedSession, error) {
	cookie, err := re.Cookies.GetCookie(r, session.AppSessionTokenCookieDef)
	if err != nil {
		// No session cookie. Simply proceed.
		return nil, nil
	}

	aSession, err := re.AppSessions.GetAppSession(ctx, HashToken(cookie.Value))
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	offlineGrant, err := re.OfflineGrantService.GetOfflineGrant(ctx, aSession.OfflineGrantID)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	}

	offlineGrantSession, ok := offlineGrant.ToSession(aSession.RefreshTokenHash)
	if !ok {
		return nil, session.ErrInvalidSession
	}

	authz, err := re.Authorizations.GetByID(ctx, offlineGrantSession.AuthorizationID)
	if errors.Is(err, ErrAuthorizationNotFound) {
		// Authorization does not exists (e.g. revoked)
		return nil, session.ErrInvalidSession
	} else if err != nil {
		return nil, err
	} else if !authz.IsAuthorized([]string{FullAccessScope}) {
		// App sessions must have full user access to be valid
		return nil, session.ErrInvalidSession
	}

	event := access.NewEvent(re.Clock.NowUTC(), re.RemoteIP, re.UserAgentString)
	offlineGrant, err = re.accessOfflineGrant(ctx, offlineGrant, aSession.RefreshTokenHash, event)
	if err != nil {
		return nil, err
	}
	offlineGrantSession, ok = offlineGrant.ToSession(aSession.RefreshTokenHash)
	if !ok {
		// This should never fail as it was a success above, so it is a panic
		panic("unexpected: invalid refresh token hash")
	}

	return offlineGrantSession, nil
}

func (re *Resolver) accessOfflineGrant(ctx context.Context, offlineGrant *OfflineGrant, refreshTokenHash string, accessEvent access.Event) (*OfflineGrant, error) {
	// When accessing the offline grant, also access its idp session
	// Access the idp session first, since the idp session expiry will be updated
	// sso enabled offline grant expiry depends on its idp session
	if offlineGrant.SSOEnabled {
		if offlineGrant.IDPSessionID == "" {
			return nil, session.ErrInvalidSession
		}
		_, err := re.Sessions.AccessWithID(ctx, offlineGrant.IDPSessionID, accessEvent)
		if errors.Is(err, idpsession.ErrSessionNotFound) {
			return nil, session.ErrInvalidSession
		} else if err != nil {
			return nil, err
		}
	}

	offlineGrant, err := re.OfflineGrantService.AccessOfflineGrant(ctx, offlineGrant.ID, refreshTokenHash, &accessEvent, offlineGrant.ExpireAtForResolvedSession)
	if err != nil {
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
