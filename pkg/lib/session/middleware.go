package session

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrInvalidSession = errors.New("provided session is invalid")

type Resolver interface {
	Resolve(rw http.ResponseWriter, r *http.Request) (Session, error)
}

type IDPSessionResolver Resolver
type AccessTokenSessionResolver Resolver

type MiddlewareLogger struct{ *log.Logger }

func NewMiddlewareLogger(lf *log.Factory) MiddlewareLogger {
	return MiddlewareLogger{lf.New("session-middleware")}
}

type AnalyticService interface {
	TrackActiveUser(userID string) error
}

type Middleware struct {
	SessionCookie              CookieDef
	Cookies                    CookieManager
	IDPSessionResolver         IDPSessionResolver
	AccessTokenSessionResolver AccessTokenSessionResolver
	AccessEvents               *access.EventProvider
	Users                      UserQuery
	Database                   *appdb.Handle
	Logger                     MiddlewareLogger
	AnalyticService            AnalyticService
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s, err := m.resolve(rw, r)

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

func (m *Middleware) resolve(rw http.ResponseWriter, r *http.Request) (s Session, err error) {
	err = m.Database.ReadOnly(func() (err error) {
		s, err = m.resolveSession(rw, r)
		if err != nil {
			return
		}
		// No session credentials provided, return no error and no resolved session
		if s == nil {
			return
		}

		// Standard session checking consider ErrUserNotFound and disabled as invalid.
		u, err := m.Users.GetRaw(s.GetAuthenticationInfo().UserID)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				err = ErrInvalidSession
			}
			return
		}
		if err = u.AccountStatus().Check(); err != nil {
			err = ErrInvalidSession
			return
		}

		event := s.GetAccessInfo().LastAccess
		err = m.AccessEvents.RecordAccess(s.SessionID(), &event)
		if err != nil {
			return
		}

		err = m.AnalyticService.TrackActiveUser(u.ID)
		if err != nil {
			return
		}

		return
	})
	return
}

func (m *Middleware) resolveSession(rw http.ResponseWriter, r *http.Request) (Session, error) {
	isInvalid := false

	// IDP session in cookie takes priority over access token in header
	for _, resolver := range []Resolver{m.IDPSessionResolver, m.AccessTokenSessionResolver} {
		session, err := resolver.Resolve(rw, r)
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
