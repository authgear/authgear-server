//+build wireinject

package handler

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideLoginHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	f LoginInteractionFlow,
	tx db.TxContext,
) http.Handler {
	h := &LoginHandler{
		Validator:    v,
		Interactions: f,
		TxContext:    tx,
	}
	return requireAuthz(h, h)
}

func newLoginHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(LoginInteractionFlow), new(*interactionflows.AuthAPIFlow)),
		provideLoginHandler,
	)
	return nil
}

func provideSignupHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	f SignupInteractionFlow,
	tx db.TxContext,
) http.Handler {
	h := &SignupHandler{
		Validator:    v,
		Interactions: f,
		TxContext:    tx,
	}
	return requireAuthz(h, h)
}

func newSignupHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(SignupInteractionFlow), new(*interactionflows.AuthAPIFlow)),
		provideSignupHandler,
	)
	return nil
}

func provideLogoutHandler(
	requireAuthz handler.RequireAuthz,
	sm logoutSessionManager,
	tx db.TxContext,
) http.Handler {
	h := &LogoutHandler{
		SessionManager: sm,
		TxContext:      tx,
	}
	return requireAuthz(h, h)
}

func newLogoutHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(logoutSessionManager), new(*auth.SessionManager)),
		provideLogoutHandler,
	)
	return nil
}

func provideRefreshHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	rp refreshProvider,
	tx db.TxContext,
) http.Handler {
	h := &RefreshHandler{
		validator:       v,
		refreshProvider: rp,
		txContext:       tx,
	}
	return requireAuthz(h, h)
}

func newRefreshHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(refreshProvider), new(*oauthhandler.TokenHandler)),
		provideRefreshHandler,
	)
	return nil
}
