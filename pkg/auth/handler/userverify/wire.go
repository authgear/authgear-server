//+build wireinject

package userverify

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

func provideUpdateHandler(requireAuthz handler.RequireAuthz, h *UpdateHandler) http.Handler {
	return requireAuthz(h, h)
}

func newUpdateHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Struct(new(UpdateHandler), "*"),
		wire.Bind(new(LoginIDProvider), new(*loginid.Provider)),
		provideUpdateHandler,
	)
	return nil
}

func provideVerifyCodeHandler(requireAuthz handler.RequireAuthz, h *VerifyCodeHandler) http.Handler {
	return requireAuthz(h, h)
}

func newVerifyCodeHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Struct(new(VerifyCodeHandler), "*"),
		wire.Bind(new(LoginIDProvider), new(*loginid.Provider)),
		provideVerifyCodeHandler,
	)
	return nil
}

func provideVerifyCodeFormHandler(h *VerifyCodeFormHandler) http.Handler {
	return h
}

func newVerifyCodeFormHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Struct(new(VerifyCodeFormHandler), "*"),
		wire.Bind(new(LoginIDProvider), new(*loginid.Provider)),
		provideVerifyCodeFormHandler,
	)
	return nil
}

func provideVerifyRequestHandler(requireAuthz handler.RequireAuthz, h *VerifyRequestHandler) http.Handler {
	return requireAuthz(h, h)
}

func newVerifyRequestHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Struct(new(VerifyRequestHandler), "*"),
		wire.Bind(new(LoginIDProvider), new(*loginid.Provider)),
		provideVerifyRequestHandler,
	)
	return nil
}
