package webapp

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SessionMiddlewareOAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
}

type SessionMiddlewareSAMLSessionService interface {
	Get(ctx context.Context, sessionID string) (*samlsession.SAMLSession, error)
}

type SessionMiddlewareSessionService interface {
	CreateSession(ctx context.Context, session *Session, redirectURI string) (*Result, error)
}

type SessionMiddlewareStore interface {
	Get(ctx context.Context, id string) (*Session, error)
}

type SessionMiddlewareOAuthUIInfoResolver interface {
	GetOAuthSessionID(req *http.Request, urlQuery string) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(ctx context.Context, r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type SessionMiddlewareSAMLUIInfoResolver interface {
	GetSAMLSessionID(req *http.Request, urlQuery string) (string, bool)
	RemoveSAMLSessionID(w http.ResponseWriter, r *http.Request)
}

type SessionMiddleware struct {
	Sessions            SessionMiddlewareSessionService
	OAuthSessions       SessionMiddlewareOAuthSessionService
	SAMLSessions        SessionMiddlewareSAMLSessionService
	States              SessionMiddlewareStore
	OAuthUIInfoResolver SessionMiddlewareOAuthUIInfoResolver
	SAMLUIInfoResolver  SessionMiddlewareSAMLUIInfoResolver
	CookieDef           SessionCookieDef
	Cookies             CookieManager
}

func (m *SessionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The session is either created now, or read from cookie.
		ctx := r.Context()

		// Create the session now.
		if oauthSessionID, ok := m.OAuthUIInfoResolver.GetOAuthSessionID(r, ""); ok {
			cookies := m.createSessionFromOAuthSession(ctx, oauthSessionID)

			// Remove oauth session ID so that we do not create again.
			m.OAuthUIInfoResolver.RemoveOAuthSessionID(w, r)

			// Redirect once to clear the oauth session id from query
			u := url.URL{
				Path:     r.URL.Path,
				RawQuery: r.URL.Query().Encode(),
			}
			redirect := httputil.ResultRedirect{
				URL:     u.String(),
				Cookies: cookies,
			}
			redirect.WriteResponse(w, r)
			return
		} else if samlSessionID, ok := m.SAMLUIInfoResolver.GetSAMLSessionID(r, ""); ok {
			cookies := m.createSessionFromSAMLSession(ctx, samlSessionID)

			// Remove saml session ID so that we do not create again.
			m.SAMLUIInfoResolver.RemoveSAMLSessionID(w, r)

			// Redirect once to clear the saml session id from query
			u := url.URL{
				Path:     r.URL.Path,
				RawQuery: r.URL.Query().Encode(),
			}
			redirect := httputil.ResultRedirect{
				URL:     u.String(),
				Cookies: cookies,
			}
			redirect.WriteResponse(w, r)
			return
		} else {
			// Or read from cookie
			ctx, session, err := m.loadSession(ctx, r)
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
				r = r.WithContext(WithSession(ctx, session))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *SessionMiddleware) createSessionFromOAuthSession(ctx context.Context, oauthSessionID string) []*http.Cookie {
	// When oauth session is not found, we fall back gracefully
	// with a zero value of SessionOptions
	sessionOptions := SessionOptions{}

	entry, err := m.OAuthSessions.Get(ctx, oauthSessionID)
	if err != nil && !errors.Is(err, oauthsession.ErrNotFound) {
		panic(err)
	}
	// err == nil || err == oauthsession.ErrNotFound

	if entry != nil {
		req := entry.T.AuthorizationRequest
		uiInfo, err := m.OAuthUIInfoResolver.ResolveForUI(ctx, req)
		if err != nil {
			panic(err)
		}
		sessionOptions = SessionOptions{
			OAuthSessionID:             oauthSessionID,
			ClientID:                   uiInfo.ClientID,
			RedirectURI:                uiInfo.RedirectURI,
			Prompt:                     uiInfo.Prompt,
			UserIDHint:                 uiInfo.UserIDHint,
			CanUseIntentReauthenticate: uiInfo.CanUseIntentReauthenticate,
			Page:                       uiInfo.Page,
			SuppressIDPSessionCookie:   uiInfo.SuppressIDPSessionCookie,
			OAuthProviderAlias:         uiInfo.OAuthProviderAlias,
			LoginHint:                  uiInfo.LoginHint,
		}

		key := otelauthgear.AttributeKeyClientID
		val := key.String(uiInfo.ClientID)
		ctx = context.WithValue(ctx, key, val)
	}

	session := NewSession(sessionOptions)

	// We do not need to redirect here so redirectURI is unimportant.
	unimportant := ""
	result, err := m.Sessions.CreateSession(ctx, session, unimportant)
	if err != nil {
		panic(err)
	}

	return result.Cookies
}

func (m *SessionMiddleware) createSessionFromSAMLSession(ctx context.Context, samlSessionID string) []*http.Cookie {
	// When saml session is not found, we fall back gracefully
	// with a zero value of SessionOptions
	sessionOptions := SessionOptions{}

	samlSession, err := m.SAMLSessions.Get(ctx, samlSessionID)
	if err != nil && !errors.Is(err, samlsession.ErrNotFound) {
		panic(err)
	}
	// err == nil || err == samlsession.ErrNotFound

	if samlSession != nil {
		uiInfo := samlSession.UIInfo
		sessionOptions = SessionOptions{
			SAMLSessionID: samlSession.ID,
			ClientID:      uiInfo.SAMLServiceProviderID,
			RedirectURI:   uiInfo.RedirectURI,
			Prompt:        uiInfo.Prompt,
			LoginHint:     uiInfo.LoginHint,
		}

		key := otelauthgear.AttributeKeyClientID
		val := key.String(uiInfo.SAMLServiceProviderID)
		ctx = context.WithValue(ctx, key, val)
	}

	session := NewSession(sessionOptions)

	// We do not need to redirect here so redirectURI is unimportant.
	unimportant := ""
	result, err := m.Sessions.CreateSession(ctx, session, unimportant)
	if err != nil {
		panic(err)
	}

	return result.Cookies
}

func (m *SessionMiddleware) loadSession(ctx context.Context, r *http.Request) (context.Context, *Session, error) {
	cookie, err := m.Cookies.GetCookie(r, m.CookieDef.Def)
	if err != nil {
		return ctx, nil, ErrSessionNotFound
	}

	s, err := m.States.Get(ctx, cookie.Value)
	if err != nil {
		return ctx, nil, err
	}

	if s.ClientID != "" {
		key := otelauthgear.AttributeKeyClientID
		val := key.String(s.ClientID)
		ctx = context.WithValue(ctx, key, val)
	}

	return ctx, s, nil
}
