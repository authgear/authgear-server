package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var ErrInvalidSession = errors.New("provided session is invalid")

type Resolver interface {
	Resolve(ctx context.Context, rw http.ResponseWriter, r *http.Request) (ResolvedSession, error)
}

type IDPSessionResolver Resolver
type AccessTokenSessionResolver Resolver

type MeterService interface {
	TrackActiveUser(ctx context.Context, userID string) error
}

type Middleware struct {
	SessionCookie              CookieDef
	Cookies                    CookieManager
	IDPSessionResolver         IDPSessionResolver
	AccessTokenSessionResolver AccessTokenSessionResolver
	AccessEvents               *access.EventProvider
	Users                      UserQuery
	Database                   *appdb.Handle
	MeterService               MeterService
	Clock                      clock.Clock
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s, err := m.resolve(r.Context(), rw, r)

		if errors.Is(err, ErrInvalidSession) {
			// Clear invalid session cookie if exist
			if _, err := m.Cookies.GetCookie(r, m.SessionCookie.Def); err == nil {
				cookie := m.Cookies.ClearCookie(m.SessionCookie.Def)
				httputil.UpdateCookie(rw, cookie)
			}

			r = r.WithContext(WithInvalidSession(r.Context()))
		} else if err != nil {
			panic(err)
		} else if s != nil {
			r = r.WithContext(WithSession(r.Context(), s))
		}
		// s is nil: no session credentials provided

		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) resolve(ctx context.Context, rw http.ResponseWriter, r *http.Request) (s ResolvedSession, err error) {
	err = m.Database.ReadOnly(ctx, func(ctx context.Context) (err error) {
		s, err = m.resolveSession(ctx, rw, r)
		if err != nil {
			return
		}
		// No session credentials provided, return no error and no resolved session
		if s == nil {
			return
		}

		// Standard session checking consider ErrUserNotFound and disabled as invalid.
		u, err := m.Users.GetRaw(ctx, s.GetAuthenticationInfo().UserID)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				err = ErrInvalidSession
			}
			return
		}
		now := m.Clock.NowUTC()
		if err = u.AccountStatus(now).Check(); err != nil {
			err = ErrInvalidSession
			return
		}

		return
	})
	return
}

func (m *Middleware) resolveSession(ctx context.Context, rw http.ResponseWriter, r *http.Request) (ResolvedSession, error) {
	isInvalid := false

	// Access token in header/App session token in cookie takes priority over IDP session in cookie
	// If both the app session and IDP session exist in the cookie
	// Middleware will read the app session first, so SDK will always open the correct settings page
	resolvers := []Resolver{m.AccessTokenSessionResolver, m.IDPSessionResolver}

	for _, resolver := range resolvers {
		session, err := resolver.Resolve(ctx, rw, r)
		if errors.Is(err, ErrInvalidSession) {
			// Continue to attempt resolving session, even if one of the resolver reported invalid.
			isInvalid = true
		} else if err != nil {
			return nil, err
		} else if session != nil {
			return session, nil
		}
	}

	if isInvalid {
		return nil, ErrInvalidSession
	}
	return nil, nil
}
