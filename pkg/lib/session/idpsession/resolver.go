package idpsession

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type resolverProvider interface {
	AccessWithToken(token string, accessEvent access.Event) (*IDPSession, error)
}

type Resolver struct {
	Cookie     session.CookieDef
	Provider   resolverProvider
	TrustProxy config.TrustProxy
	Clock      clock.Clock
}

func (re *Resolver) Resolve(rw http.ResponseWriter, r *http.Request) (session.Session, error) {
	cookie, err := r.Cookie(re.Cookie.Def.Name)
	if err != nil {
		// No cookie. Simply proceed.
		return nil, nil
	}

	accessEvent := access.NewEvent(re.Clock.NowUTC(), r, bool(re.TrustProxy))
	s, err := re.Provider.AccessWithToken(cookie.Value, accessEvent)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			err = session.ErrInvalidSession
		}
		return nil, err
	}

	return s, nil
}
