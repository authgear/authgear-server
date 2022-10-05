package admin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/admin/facade"
	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/admin/service"
	"github.com/authgear/authgear-server/pkg/admin/transport"
	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/event"
	libfacade "github.com/authgear/authgear-server/pkg/lib/facade"
	featurecustomattrs "github.com/authgear/authgear-server/pkg/lib/feature/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	featurestdattrs "github.com/authgear/authgear-server/pkg/lib/feature/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/nonce"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/presign"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	middleware.DependencySet,

	nonce.DependencySet,
	wire.Bind(new(interaction.NonceService), new(*nonce.Service)),

	loader.DependencySet,
	wire.Bind(new(loader.UserLoaderUserService), new(*user.Queries)),
	wire.Bind(new(loader.IdentityLoaderIdentityService), new(*identityservice.Service)),
	wire.Bind(new(loader.AuthenticatorLoaderAuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(loader.AuditLogQuery), new(*audit.Query)),

	facade.DependencySet,
	wire.Bind(new(facade.UserService), new(*libfacade.UserFacade)),
	wire.Bind(new(facade.UserSearchService), new(*libes.Service)),
	wire.Bind(new(facade.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(facade.AuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(facade.InteractionService), new(*service.InteractionService)),
	wire.Bind(new(facade.VerificationService), new(*verification.Service)),
	wire.Bind(new(facade.StandardAttributesService), new(*featurestdattrs.ServiceNoEvent)),
	wire.Bind(new(facade.CustomAttributesService), new(*featurecustomattrs.ServiceNoEvent)),
	wire.Bind(new(facade.SessionManager), new(*session.Manager)),
	wire.Bind(new(facade.AuditLogQuery), new(*audit.Query)),
	wire.Bind(new(facade.EventService), new(*event.Service)),
	wire.Bind(new(facade.AuthorizationService), new(*oauth.AuthorizationService)),

	graphql.DependencySet,
	wire.Bind(new(graphql.UserLoader), new(*loader.UserLoader)),
	wire.Bind(new(graphql.IdentityLoader), new(*loader.IdentityLoader)),
	wire.Bind(new(graphql.AuthenticatorLoader), new(*loader.AuthenticatorLoader)),
	wire.Bind(new(graphql.AuditLogLoader), new(*loader.AuditLogLoader)),
	wire.Bind(new(graphql.UserFacade), new(*facade.UserFacade)),
	wire.Bind(new(graphql.IdentityFacade), new(*facade.IdentityFacade)),
	wire.Bind(new(graphql.AuthenticatorFacade), new(*facade.AuthenticatorFacade)),
	wire.Bind(new(graphql.VerificationFacade), new(*facade.VerificationFacade)),
	wire.Bind(new(graphql.SessionFacade), new(*facade.SessionFacade)),
	wire.Bind(new(graphql.AuditLogFacade), new(*facade.AuditLogFacade)),
	wire.Bind(new(graphql.UserProfileFacade), new(*facade.UserProfileFacade)),
	wire.Bind(new(graphql.AuthorizationFacade), new(*facade.AuthorizationFacade)),

	service.DependencySet,
	wire.Bind(new(service.InteractionGraphService), new(*interaction.Service)),

	wire.Struct(new(WebEndpoints), "*"),
	wire.Bind(new(sso.EndpointsProvider), new(*WebEndpoints)),
	wire.Bind(new(sso.RedirectURLProvider), new(*WebEndpoints)),
	wire.Bind(new(otp.EndpointsProvider), new(*WebEndpoints)),
	wire.Bind(new(forgotpassword.URLProvider), new(*WebEndpoints)),
	wire.Bind(new(sso.WechatURLProvider), new(*WebEndpoints)),

	transport.DependencySet,
	wire.Bind(new(transport.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(transport.PresignProvider), new(*presign.Provider)),

	adminauthz.DependencySet,
)
