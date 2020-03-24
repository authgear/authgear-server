package session

import (
	"errors"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
)

type ResolverProvider interface {
	GetByToken(token string) (*IDPSession, error)
	Update(session *IDPSession) error
}

type Resolver struct {
	CookieConfiguration CookieConfiguration
	Provider            ResolverProvider
}

func (re *Resolver) Resolve(rw http.ResponseWriter, r *http.Request) (auth.AuthSession, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		// No cookie. Simply proceed.
		return nil, nil
	}

	session, err := re.Provider.GetByToken(cookie.Value)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			err = auth.ErrInvalidSession
			re.CookieConfiguration.Clear(rw)
		}
		return nil, err
	}

	return session, nil
}

func (re *Resolver) OnAccess(session auth.AuthSession, event auth.AccessEvent) error {
	s := session.(*IDPSession)
	s.AccessInfo.LastAccess = event
	return re.Provider.Update(s)
}
