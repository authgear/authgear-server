//+build wireinject

package forgotpwd

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpassword"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideForgotPasswordHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	fpp ForgotPasswordProvider,
	tx db.TxContext,
) http.Handler {
	h := &ForgotPasswordHandler{
		Validator:              v,
		ForgotPasswordProvider: fpp,
		TxContext:              tx,
	}
	return requireAuthz(h, h)
}

func newForgotPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(ForgotPasswordProvider), new(*forgotpassword.Provider)),
		provideForgotPasswordHandler,
	)
	return nil
}

func provideResetPasswordHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	rpp ResetPasswordProvider,
	tx db.TxContext,
) http.Handler {
	h := &ResetPasswordHandler{
		Validator:             v,
		ResetPasswordProvider: rpp,
		TxContext:             tx,
	}
	return requireAuthz(h, h)
}

func newResetPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(ResetPasswordProvider), new(*forgotpassword.Provider)),
		provideResetPasswordHandler,
	)
	return nil
}
