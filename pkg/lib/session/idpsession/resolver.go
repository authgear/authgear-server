package idpsession

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type resolverProvider interface {
	GetByToken(token string) (*IDPSession, error)
	Update(session *IDPSession) error
}

type Resolver struct {
	CookieFactory CookieFactory
	Cookie        CookieDef
	Provider      resolverProvider
	Config        *config.ServerConfig
	Clock         clock.Clock
}

func (re *Resolver) Resolve(rw http.ResponseWriter, r *http.Request) (session.Session, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		// No cookie. Simply proceed.
		return nil, nil
	}

	s, err := re.Provider.GetByToken(cookie.Value)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			err = session.ErrInvalidSession
			cookie := re.CookieFactory.ClearCookie(re.Cookie.Def)
			httputil.UpdateCookie(rw, cookie)
		}
		return nil, err
	}

	s.AccessInfo.LastAccess = access.NewEvent(re.Clock.NowUTC(), r, re.Config.TrustProxy)
	if err = re.Provider.Update(s); err != nil {
		return nil, err
	}

	return s, nil
}
