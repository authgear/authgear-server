//go:build wireinject
// +build wireinject

package auth

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/api"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newWebAppRequestMiddleware(w http.ResponseWriter, r *http.Request, p *deps.RootProvider, configSource *configsource.ConfigSource) httproute.Middleware {
	panic(wire.Build(
		RequestMiddlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*WebAppRequestMiddleware)),
	))
}

func newPanicMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.PanicMiddleware)),
	))
}

func newOtelMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		otelauthgear.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*otelauthgear.HTTPInstrumentationMiddleware)),
	))
}

func newSentryMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newBodyLimitMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.BodyLimitMiddleware)),
	))
}

func newPanicWebAppMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*handlerwebapp.PanicMiddleware)),
	))
}

func newPublicOriginMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.PublicOriginMiddleware)),
	))
}

func newCORSMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.CORSMiddleware)),
	))
}

func newContextHolderMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		webapp.RootMiddlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.ContextHolderMiddleware)),
	))
}

func newDynamicCSPMiddleware(
	p *deps.RequestProvider,
	allowFrameAncestorsFromEnv webapp.AllowFrameAncestorsFromEnv,
	allowFrameAncestorsFromCustomUI webapp.AllowFrameAncestorsFromCustomUI,
) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.DynamicCSPMiddleware)),
	))
}

func newNoProjectCSPMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		NoProjectDependencySet,
		webapp.RootMiddlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.NoProjectCSPMiddleware)),
	))
}

func newCSRFMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*handlerwebapp.CSRFMiddleware)),
	))
}

func newCSRFDebugMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.CSRFDebugMiddleware)),
	))
}

func newAuthEntryPointMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*handlerwebapp.AuthEntryPointMiddleware)),
	))
}

func newSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*session.Middleware)),
	))
}

func newWebAppSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.SessionMiddleware)),
	))
}

func newWebAppColorSchemeMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.ColorSchemeMiddleware)),
	))
}

func newWebAppUIParamMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.UIParamMiddleware)),
	))
}

func newWebAppWeChatRedirectURIMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.WeChatRedirectURIMiddleware)),
	))
}

func newWebAppVisitorIDMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.VisitorIDMiddleware)),
	))
}

func newRequireAuthenticationEnabledMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.RequireAuthenticationEnabledMiddleware)),
	))
}

func newRequireSettingsEnabledMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.RequireSettingsEnabledMiddleware)),
	))
}

func newSettingsSubRoutesMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.SettingsSubRoutesMiddleware)),
	))
}

func newSuccessPageMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.SuccessPageMiddleware)),
	))
}

func newAPIRRequireAuthenticatedMiddlewareMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*api.RequireAuthenticatedMiddleware)),
	))
}

func newTutorialMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.TutorialMiddleware)),
	))
}

func newWorkflowIntlMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*workflow.IntlMiddleware)),
	))
}

func newAuthenticationFlowIntlMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*authenticationflow.IntlMiddleware)),
	))
}

func newAuthenticationFlowRateLimitMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*authenticationflow.RateLimitMiddleware)),
	))
}

func newAccountManagementRateLimitMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*accountmanagement.RateLimitMiddleware)),
	))
}

func newImplementationSwitcherMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*handlerwebapp.ImplementationSwitcherMiddleware)),
	))
}

func newSettingImplementationSwitcherMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*handlerwebapp.SettingsImplementationSwitcherMiddleware)),
	))
}

func newDPoPMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*dpop.Middleware)),
	))
}
