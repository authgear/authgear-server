package auth

import (
	"errors"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

var ErrInvalidSession = errors.New("provided session is invalid")

type IDPSessionResolver interface {
	Resolve(rw http.ResponseWriter, r *http.Request) (AuthSession, error)
	OnAccess(session AuthSession, event AccessEvent) error
}

type Middleware struct {
	IDPSessionResolver IDPSessionResolver
	AuthInfoStore      authinfo.Store
	TxContext          db.TxContext
	Time               time.Provider
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s, u, err := m.resolve(rw, r)

		if errors.Is(err, ErrInvalidSession) {
			r = r.WithContext(authn.WithInvalidAuthn(r.Context()))
		} else if err != nil {
			panic(err)
		} else if s != nil {
			r = r.WithContext(authn.WithAuthn(r.Context(), s, u))
		}
		// s is nil: no session credentials provided

		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) resolve(rw http.ResponseWriter, r *http.Request) (AuthSession, *authinfo.AuthInfo, error) {
	if err := m.TxContext.BeginTx(); err != nil {
		return nil, nil, err
	}
	defer m.TxContext.RollbackTx()

	sessionIDP, err := m.IDPSessionResolver.Resolve(rw, r)
	if err != nil {
		return nil, nil, err
	}

	if sessionIDP == nil {
		return nil, nil, nil
	}

	user := &authinfo.AuthInfo{}
	if err = m.AuthInfoStore.GetAuth(sessionIDP.AuthnAttrs().UserID, user); err != nil {
		return nil, nil, err
	}

	accessEvent := NewAccessEvent(m.Time.NowUTC(), r)
	err = m.IDPSessionResolver.OnAccess(sessionIDP, accessEvent)
	if err != nil {
		return nil, nil, err
	}

	return sessionIDP, user, nil
}
