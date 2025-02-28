package idpsession

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type resolverProvider interface {
	AccessWithToken(ctx context.Context, token string, accessEvent access.Event) (*IDPSession, error)
}

type ResolverCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
}

type Resolver struct {
	Cookies         ResolverCookieManager
	CookieDef       session.CookieDef
	Provider        resolverProvider
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	TrustProxy      config.TrustProxy
	Clock           clock.Clock
}

func (re *Resolver) Resolve(ctx context.Context, rw http.ResponseWriter, r *http.Request) (session.ResolvedSession, error) {
	cookie, err := re.Cookies.GetCookie(r, re.CookieDef.Def)
	if err != nil {
		// No cookie. Simply proceed.
		return nil, nil
	}

	accessEvent := access.NewEvent(re.Clock.NowUTC(), re.RemoteIP, re.UserAgentString)
	s, err := re.Provider.AccessWithToken(ctx, cookie.Value, accessEvent)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			err = session.ErrInvalidSession
		}
		return nil, err
	}

	return s, nil
}
