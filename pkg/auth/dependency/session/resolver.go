package session

import (
	"errors"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type ResolverProvider interface {
	GetByToken(token string) (*IDPSession, error)
	Update(session *IDPSession) error
}

type Resolver struct {
	CookieConfiguration CookieConfiguration
	Provider            ResolverProvider
	Time                time.Provider
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

	session.AccessInfo.LastAccess = auth.NewAccessEvent(re.Time.NowUTC(), r)
	if err = re.Provider.Update(session); err != nil {
		return nil, err
	}

	return session, nil
}
