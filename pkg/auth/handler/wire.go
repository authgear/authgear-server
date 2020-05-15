//+build wireinject

package handler

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	identityprovider "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/provider"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
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

func provideChangePasswordHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	as authinfo.Store,
	tx db.TxContext,
	ups userprofile.Store,
	hp hook.Provider,
	aq async.Queue,
	f PasswordFlow,
) http.Handler {
	h := &ChangePasswordHandler{
		Validator:        v,
		AuthInfoStore:    as,
		TxContext:        tx,
		UserProfileStore: ups,
		HookProvider:     hp,
		TaskQueue:        aq,
		Interactions:     f,
	}
	return requireAuthz(h, h)
}

func newChangePasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(PasswordFlow), new(*interactionflows.PasswordFlow)),
		provideChangePasswordHandler,
	)
	return nil
}

func provideResetPasswordHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	ups userprofile.Store,
	as authinfo.Store,
	tx db.TxContext,
	aq async.Queue,
	hp hook.Provider,
	f ResetPasswordFlow,
) http.Handler {
	h := &ResetPasswordHandler{
		Validator:        v,
		UserProfileStore: ups,
		AuthInfoStore:    as,
		TxContext:        tx,
		TaskQueue:        aq,
		HookProvider:     hp,
		Interactions:     f,
	}
	return requireAuthz(h, h)
}

func newResetPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(ResetPasswordFlow), new(*interactionflows.PasswordFlow)),
		provideResetPasswordHandler,
	)
	return nil
}

func providerListIdentitiesHandler(
	requireAuthz handler.RequireAuthz,
	tx db.TxContext,
	ip ListIdentityProvider,
) http.Handler {
	h := &ListIdentitiesHandler{
		TxContext:        tx,
		IdentityProvider: ip,
	}
	return requireAuthz(h, h)
}

func newListIdentitiesHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(ListIdentityProvider), new(*identityprovider.Provider)),
		providerListIdentitiesHandler,
	)
	return nil
}
