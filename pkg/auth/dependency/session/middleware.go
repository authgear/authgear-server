package session

import (
	"errors"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Middleware struct {
	UseInsecureCookie bool
	SessionProvider   Provider
	AuthInfoStore     authinfo.Store
	TxContext         db.TxContext
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s, u, err := m.resolve(r)

		if errors.Is(err, session.ErrSessionNotFound) {
			ClearCookie(rw, m.UseInsecureCookie)
		} else if err != nil {
			panic(err)
		}

		r = r.WithContext(WithSession(r.Context(), s, u))
		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) resolve(r *http.Request) (*Session, *authinfo.AuthInfo, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		// No cookie. Simply proceed.
		return nil, nil, nil
	}

	if err = m.TxContext.BeginTx(); err != nil {
		return nil, nil, err
	}
	defer m.TxContext.RollbackTx()

	token := cookie.Value
	session, err := m.SessionProvider.Get(token)
	if err != nil {
		return nil, nil, err
	}

	err = m.SessionProvider.Access(session)
	if err != nil {
		return nil, nil, err
	}

	user := &authinfo.AuthInfo{}
	if err = m.AuthInfoStore.GetAuth(session.UserID, user); err != nil {
		return nil, nil, err
	}

	return session, user, nil
}
