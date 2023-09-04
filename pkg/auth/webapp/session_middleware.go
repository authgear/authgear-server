package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SessionMiddlewareOAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type SessionMiddlewareSessionService interface {
	CreateSession(session *Session, redirectURI string) (*Result, error)
}

type SessionMiddlewareStore interface {
	Get(id string) (*Session, error)
}

type SessionMiddlewareUIInfoResolver interface {
	GetOAuthSessionID(req *http.Request) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type SessionMiddleware struct {
	Sessions       SessionMiddlewareSessionService
	OAuthSessions  SessionMiddlewareOAuthSessionService
	States         SessionMiddlewareStore
	UIInfoResolver SessionMiddlewareUIInfoResolver
	CookieDef      SessionCookieDef
	Cookies        CookieManager
}

func (m *SessionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The session is either created now, or read from cookie.

		// Create the session now.
		if oauthSessionID, ok := m.UIInfoResolver.GetOAuthSessionID(r); ok {
			result, session := m.createSession(oauthSessionID)

			for _, c := range result.Cookies {
				httputil.UpdateCookie(w, c)
			}

			// Remove oauth session ID so that we do not create again.
			m.UIInfoResolver.RemoveOAuthSessionID(w, r)

			r = r.WithContext(WithSession(r.Context(), session))
		} else {
			// Or read from cookie
			session, err := m.loadSession(r)
			if err != nil {
				if errors.Is(err, ErrSessionNotFound) {
					// fallthrough
				} else if errors.Is(err, ErrInvalidSession) {
					// Clear the session before continuing
					cookie := m.Cookies.ClearCookie(m.CookieDef.Def)
					httputil.UpdateCookie(w, cookie)
				} else {
					panic(err)
				}
			}
			if session != nil {
				r = r.WithContext(WithSession(r.Context(), session))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *SessionMiddleware) createSession(oauthSessionID string) (*Result, *Session) {
	// When oauth session is not found, we fall back gracefully
	// with a zero value of SessionOptions
	sessionOptions := SessionOptions{}

	entry, err := m.OAuthSessions.Get(oauthSessionID)
	if err != nil && !errors.Is(err, oauthsession.ErrNotFound) {
		panic(err)
	}
	// err == nil || err == oauthsession.ErrNotFound

	if entry != nil {
		req := entry.T.AuthorizationRequest
		uiInfo, err := m.UIInfoResolver.ResolveForUI(req)
		if err != nil {
			panic(err)
		}
		sessionOptions = SessionOptions{
			OAuthSessionID:             oauthSessionID,
			RedirectURI:                uiInfo.RedirectURI,
			Prompt:                     uiInfo.Prompt,
			UserIDHint:                 uiInfo.UserIDHint,
			CanUseIntentReauthenticate: uiInfo.CanUseIntentReauthenticate,
			Page:                       uiInfo.Page,
			SuppressIDPSessionCookie:   uiInfo.SuppressIDPSessionCookie,
			OAuthProviderAlias:         uiInfo.OAuthProviderAlias,
			LoginHint:                  uiInfo.LoginHint,
		}
	}

	session := NewSession(sessionOptions)

	// We do not need to redirect here so redirectURI is unimportant.
	unimportant := ""
	result, err := m.Sessions.CreateSession(session, unimportant)
	if err != nil {
		panic(err)
	}

	return result, session
}

func (m *SessionMiddleware) loadSession(r *http.Request) (*Session, error) {
	cookie, err := m.Cookies.GetCookie(r, m.CookieDef.Def)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	s, err := m.States.Get(cookie.Value)
	if err != nil {
		return nil, err
	}

	return s, nil
}
