package auth

import (
	"errors"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

var ErrInvalidSession = errors.New("provided session is invalid")

type SessionResolver interface {
	Resolve(rw http.ResponseWriter, r *http.Request) (AuthSession, error)
}

type IDPSessionResolver SessionResolver
type AccessTokenSessionResolver SessionResolver

type Middleware struct {
	IDPSessionResolver         IDPSessionResolver
	AccessTokenSessionResolver AccessTokenSessionResolver
	AccessEvents               AccessEventProvider
	Users                      UserProvider
	TxContext                  db.TxContext
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s, u, err := m.resolve(rw, r)

		if errors.Is(err, ErrInvalidSession) {
			r = r.WithContext(authn.WithInvalidAuthn(r.Context()))
		} else if err != nil {
			panic(err)
		} else if s != nil {
			r = r.WithContext(authn.WithAuthn(r.Context(), s, u.ToUserInfo()))
		}
		// s is nil: no session credentials provided

		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) resolve(rw http.ResponseWriter, r *http.Request) (session AuthSession, user *model.User, err error) {
	err = db.ReadOnly(m.TxContext, func() (err error) {
		session, err = m.resolveSession(rw, r)
		if err != nil {
			return
		}
		// No session credentials provided, return no error and no resolved session
		if session == nil {
			return
		}
		user, err = m.Users.Get(session.AuthnAttrs().UserID)
		if err != nil {
			return
		}
		event := session.GetAccessInfo().LastAccess
		err = m.AccessEvents.RecordAccess(session, event)
		if err != nil {
			return
		}
		return
	})
	return
}

func (m *Middleware) resolveSession(rw http.ResponseWriter, r *http.Request) (AuthSession, error) {
	isInvalid := false

	// IDP session in cookie takes priority over access token in header
	for _, resolver := range []SessionResolver{m.IDPSessionResolver, m.AccessTokenSessionResolver} {
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
