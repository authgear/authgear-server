package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SessionMiddlewareStore interface {
	Get(id string) (*Session, error)
}

type SessionMiddleware struct {
	States    SessionMiddlewareStore
	CookieDef SessionCookieDef
	Cookies   CookieManager
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
			cookie := m.Cookies.ClearCookie(m.CookieDef.Def)
			httputil.UpdateCookie(w, cookie)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			panic(err)
		}

		r = r.WithContext(WithSession(r.Context(), session))

		next.ServeHTTP(w, r)
	})
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
