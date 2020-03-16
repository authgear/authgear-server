package session

import (
	"errors"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Resolver interface {
	GetByToken(token string) (*Session, error)
	Access(*Session) error
}

type Middleware struct {
	CookieConfiguration CookieConfiguration
	SessionResolver     Resolver
	AuthInfoStore       authinfo.Store
	TxContext           db.TxContext
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil {
			// No cookie. Simply proceed.
			next.ServeHTTP(rw, r)
			return
		}

		s, u, err := m.resolve(cookie.Value)

		if errors.Is(err, ErrSessionNotFound) {
			ClearCookie(rw, m.CookieConfiguration)
		} else if err != nil {
			panic(err)
		}

		r = r.WithContext(WithSession(r.Context(), s, u))
		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) resolve(token string) (*Session, *authinfo.AuthInfo, error) {
	if err := m.TxContext.BeginTx(); err != nil {
		return nil, nil, err
	}
	defer m.TxContext.RollbackTx()

	session, err := m.SessionResolver.GetByToken(token)
	if err != nil {
		return nil, nil, err
	}

	err = m.SessionResolver.Access(session)
	if err != nil {
		return nil, nil, err
	}

	user := &authinfo.AuthInfo{}
	if err = m.AuthInfoStore.GetAuth(session.UserID, user); err != nil {
		return nil, nil, err
	}

	return session, user, nil
}
