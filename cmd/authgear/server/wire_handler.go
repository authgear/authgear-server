//+build wireinject

package server

import (
	"net/http"

	"github.com/google/wire"

	handlerinternal "github.com/authgear/authgear-server/pkg/auth/handler/internalserver"
	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/deps"
)

func newSessionResolveHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerinternal.ResolveHandler)),
	))
}

func newOAuthAuthorizeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.AuthorizeHandler)),
	))
}

func newOAuthTokenHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.TokenHandler)),
	))
}

func newOAuthRevokeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.RevokeHandler)),
	))
}

func newOAuthMetadataHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.MetadataHandler)),
	))
}

func newOAuthJWKSHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.JWKSHandler)),
	))
}

func newOAuthUserInfoHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.UserInfoHandler)),
	))
}

func newOAuthEndSessionHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.EndSessionHandler)),
	))
}

func newOAuthChallengeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.ChallengeHandler)),
	))
}

func newWebAppRootHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.RootHandler)),
	))
}

func newWebAppLoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.LoginHandler)),
	))
}

func newWebAppSignupHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SignupHandler)),
	))
}

func newWebAppPromoteHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.PromoteHandler)),
	))
}

func newWebAppSSOCallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SSOCallbackHandler)),
	))
}

func newWebAppEnterLoginIDHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterLoginIDHandler)),
	))
}

func newWebAppEnterPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterPasswordHandler)),
	))
}

func newWebAppCreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.CreatePasswordHandler)),
	))
}

func newWebAppEnterOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterOOBOTPHandler)),
	))
}

func newWebAppForgotPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForgotPasswordHandler)),
	))
}

func newWebAppForgotPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForgotPasswordSuccessHandler)),
	))
}

func newWebAppResetPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ResetPasswordHandler)),
	))
}

func newWebAppResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ResetPasswordSuccessHandler)),
	))
}

func newWebAppSettingsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsHandler)),
	))
}

func newWebAppSettingsIdentityHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsIdentityHandler)),
	))
}

func newWebAppLogoutHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.LogoutHandler)),
	))
}
