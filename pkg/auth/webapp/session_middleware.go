package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type SessionMiddlewareStore interface {
	Get(id string) (*Session, error)
}

type SessionMiddleware struct {
	States        SessionMiddlewareStore
	Cookie        SessionCookieDef
	CookieFactory CookieFactory
}

func (m *SessionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.loadSession(r)
		if errors.Is(err, ErrSessionNotFound) {
			// Continue without session
			next.ServeHTTP(w, r)
			return
		} else if errors.Is(err, ErrInvalidSession) {
			// Clear the session before continuing
			cookie := m.CookieFactory.ClearCookie(m.Cookie.Def)
			httputil.UpdateCookie(w, cookie)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			panic(err)
		}

		r = r.WithContext(WithSession(r.Context(), session))

		// Restore UI locales from state if necessary
		if session.UILocales != "" {
			tags := intl.ParseUILocales(session.UILocales)
			ctx := intl.WithPreferredLanguageTags(r.Context(), tags)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (m *SessionMiddleware) loadSession(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(m.Cookie.Def.Name)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	s, err := m.States.Get(cookie.Value)
	if err != nil {
		return nil, err
	}

	return s, nil
}
