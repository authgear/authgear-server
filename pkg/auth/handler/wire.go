//+build wireinject

package handler

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideLoginHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	ap LoginAuthnProvider,
	tx db.TxContext,
) http.Handler {
	h := &LoginHandler{
		Validator:     v,
		AuthnProvider: ap,
		TxContext:     tx,
	}
	return requireAuthz(h, h)
}

func newLoginHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(LoginAuthnProvider), new(*authn.Provider)),
		provideLoginHandler,
	)
	return nil
}

func provideSignupHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	ap SignupAuthnProvider,
	tx db.TxContext,
) http.Handler {
	h := &SignupHandler{
		Validator:     v,
		AuthnProvider: ap,
		TxContext:     tx,
	}
	return requireAuthz(h, h)
}

func newSignupHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(SignupAuthnProvider), new(*authn.Provider)),
		provideSignupHandler,
	)
	return nil
}
