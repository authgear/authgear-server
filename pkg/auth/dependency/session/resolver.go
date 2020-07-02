package session

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/clock"
)

type resolverProvider interface {
	GetByToken(token string) (*IDPSession, error)
	Update(session *IDPSession) error
}

type Resolver struct {
	Cookie   CookieDef
	Provider resolverProvider
	Config   *config.ServerConfig
	Clock    clock.Clock
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
			re.Cookie.Clear(rw)
		}
		return nil, err
	}

	session.AccessInfo.LastAccess = auth.NewAccessEvent(re.Clock.NowUTC(), r, re.Config.TrustProxy)
	if err = re.Provider.Update(session); err != nil {
		return nil, err
	}

	return session, nil
}
