//+build wireinject

package session

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideResolveHandler(
	m *auth.Middleware,
	lf logging.Factory,
	t time.Provider,
	ap *anonymous.Provider,
) http.Handler {
	return m.Handle(&ResolveHandler{
		TimeProvider:  t,
		LoggerFactory: lf,
		Anonymous:     ap,
	})
}

func newResolveHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(pkg.DependencySet, provideResolveHandler)
	return nil
}

func provideListHandler(tx db.TxContext, sm sessionListManager, requireAuthz handler.RequireAuthz) http.Handler {
	h := &ListHandler{
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func newListHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(sessionListManager), new(*auth.SessionManager)),
		provideListHandler,
	)
	return nil
}

func provideGetHandler(tx db.TxContext, sm sessionGetManager, v *validation.Validator, requireAuthz handler.RequireAuthz) http.Handler {
	h := &GetHandler{
		validator:      v,
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func newGetHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(sessionGetManager), new(*auth.SessionManager)),
		provideGetHandler,
	)
	return nil
}

func provideRevokeHandler(tx db.TxContext, sm sessionRevokeManager, v *validation.Validator, requireAuthz handler.RequireAuthz) http.Handler {
	h := &RevokeHandler{
		validator:      v,
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func newRevokeHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(sessionRevokeManager), new(*auth.SessionManager)),
		provideRevokeHandler,
	)
	return nil
}

func provideRevokeAllHandler(tx db.TxContext, sm sessionRevokeAllManager, requireAuthz handler.RequireAuthz) http.Handler {
	h := &RevokeAllHandler{
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func newRevokeAllHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(sessionRevokeAllManager), new(*auth.SessionManager)),
		provideRevokeAllHandler,
	)
	return nil
}
