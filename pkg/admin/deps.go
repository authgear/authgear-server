package admin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/admin/facade"
	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/admin/service"
	"github.com/authgear/authgear-server/pkg/admin/transport"
	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	middleware.DependencySet,

	loader.DependencySet,
	wire.Bind(new(loader.UserLoaderUserService), new(*user.Queries)),
	wire.Bind(new(loader.IdentityLoaderIdentityService), new(*identityservice.Service)),
	wire.Bind(new(loader.AuthenticatorLoaderAuthenticatorService), new(*authenticatorservice.Service)),

	facade.DependencySet,
	wire.Bind(new(facade.UserService), new(*user.Provider)),
	wire.Bind(new(facade.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(facade.AuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(facade.InteractionService), new(*service.InteractionService)),
	wire.Bind(new(facade.VerificationService), new(*verification.Service)),
	wire.Bind(new(facade.SessionManager), new(*session.Manager)),

	graphql.DependencySet,
	wire.Bind(new(graphql.UserLoader), new(*loader.UserLoader)),
	wire.Bind(new(graphql.IdentityLoader), new(*loader.IdentityLoader)),
	wire.Bind(new(graphql.AuthenticatorLoader), new(*loader.AuthenticatorLoader)),
	wire.Bind(new(graphql.UserFacade), new(*facade.UserFacade)),
	wire.Bind(new(graphql.IdentityFacade), new(*facade.IdentityFacade)),
	wire.Bind(new(graphql.AuthenticatorFacade), new(*facade.AuthenticatorFacade)),
	wire.Bind(new(graphql.VerificationFacade), new(*facade.VerificationFacade)),
	wire.Bind(new(graphql.SessionFacade), new(*facade.SessionFacade)),

	service.DependencySet,
	wire.Bind(new(service.InteractionGraphService), new(*interaction.Service)),

	wire.Struct(new(WebEndpoints), "*"),
	wire.Bind(new(sso.EndpointsProvider), new(*WebEndpoints)),
	wire.Bind(new(sso.RedirectURLProvider), new(*WebEndpoints)),
	wire.Bind(new(otp.EndpointsProvider), new(*WebEndpoints)),
	wire.Bind(new(verification.WebAppURLProvider), new(*WebEndpoints)),
	wire.Bind(new(forgotpassword.URLProvider), new(*WebEndpoints)),

	transport.DependencySet,
	adminauthz.DependencySet,
)
