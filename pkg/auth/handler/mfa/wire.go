//+build wireinject

package mfa

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

var dependencySet = wire.NewSet(
	auth.DependencySet,
	wire.Bind(new(authnResolver), new(*authn.Provider)),
	wire.Bind(new(authnStepper), new(*authn.Provider)),
)

func provideActivateOOBHandler(h *ActivateOOBHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newActivateOOBHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(ActivateOOBHandler), "*"),
		provideActivateOOBHandler,
	)
	return nil
}

func provideActivateTOTPHandler(h *ActivateTOTPHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newActivateTOTPHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(ActivateTOTPHandler), "*"),
		provideActivateTOTPHandler,
	)
	return nil
}

func provideAuthenticateBearerTokenHandler(h *AuthenticateBearerTokenHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newAuthenticateBearerTokenHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(AuthenticateBearerTokenHandler), "*"),
		provideAuthenticateBearerTokenHandler,
	)
	return nil
}

func provideAuthenticateOOBHandler(h *AuthenticateOOBHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newAuthenticateOOBHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(AuthenticateOOBHandler), "*"),
		provideAuthenticateOOBHandler,
	)
	return nil
}

func provideAuthenticateRecoveryCodeHandler(h *AuthenticateRecoveryCodeHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newAuthenticateRecoveryCodeHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(AuthenticateRecoveryCodeHandler), "*"),
		provideAuthenticateRecoveryCodeHandler,
	)
	return nil
}

func provideAuthenticateTOTPHandler(h *AuthenticateTOTPHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newAuthenticateTOTPHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(AuthenticateTOTPHandler), "*"),
		provideAuthenticateTOTPHandler,
	)
	return nil
}

func provideCreateOOBHandler(h *CreateOOBHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newCreateOOBHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(CreateOOBHandler), "*"),
		provideCreateOOBHandler,
	)
	return nil
}

func provideCreateTOTPHandler(h *CreateTOTPHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newCreateTOTPHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(CreateTOTPHandler), "*"),
		provideCreateTOTPHandler,
	)
	return nil
}

func provideListAuthenticatorHandler(h *ListAuthenticatorHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newListAuthenticatorHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(ListAuthenticatorHandler), "*"),
		provideListAuthenticatorHandler,
	)
	return nil
}

func provideTriggerOOBHandler(h *TriggerOOBHandler, requireAuthz handler.RequireAuthz) http.Handler {
	return requireAuthz(h, h)
}

func newTriggerOOBHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(TriggerOOBHandler), "*"),
		provideTriggerOOBHandler,
	)
	return nil
}
