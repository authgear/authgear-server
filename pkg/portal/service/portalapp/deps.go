package portalapp

import (
	"github.com/authgear/authgear-server/pkg/lib/audit"
	authenticatoroob "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	authenticatorpasskey "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/passkey"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	authenticatortotp "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	identitypasskey "github.com/authgear/authgear-server/pkg/lib/authn/identity/passkey"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	identitysiwe "github.com/authgear/authgear-server/pkg/lib/authn/identity/siwe"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/event"
	featurecustomattrs "github.com/authgear/authgear-server/pkg/lib/feature/customattrs"
	featurepasskey "github.com/authgear/authgear-server/pkg/lib/feature/passkey"
	featuresiwe "github.com/authgear/authgear-server/pkg/lib/feature/siwe"
	featurestdattrs "github.com/authgear/authgear-server/pkg/lib/feature/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	featureweb3 "github.com/authgear/authgear-server/pkg/lib/feature/web3"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/lib/web"
	portaldeps "github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	portaldeps.AppDependencySet,

	deps.ConfigDeps,
	appdb.DependencySet,
	user.DependencySet,
	identityservice.DependencySet,
	identityloginid.DependencySet,
	identityoauth.DependencySet,
	identityanonymous.DependencySet,
	identitybiometric.DependencySet,
	identitypasskey.DependencySet,
	identitysiwe.DependencySet,
	featurepasskey.DependencySet,
	translation.DependencySet,
	template.DependencySet,
	event.DependencySet,
	clock.DependencySet,
	web.DependencySet,
	featuresiwe.DependencySet,
	ratelimit.DependencySet,
	authenticatorservice.DependencySet,
	authenticatorpassword.DependencySet,
	authenticatorpasskey.DependencySet,
	authenticatortotp.DependencySet,
	authenticatoroob.DependencySet,
	featurestdattrs.DependencySet,
	otp.DependencySet,
	verification.DependencySet,
	featurecustomattrs.DependencySet,
	featureweb3.DependencySet,
	hook.DependencySet,
	audit.DependencySet,
	tutorial.DependencySet,
	elasticsearch.DependencySet,
	auditdb.DependencySet,
	globaldb.DependencySet,
	wire.Bind(new(hook.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(hook.StandardAttributesServiceNoEvent), new(*featurestdattrs.ServiceNoEvent)),
	wire.Bind(new(hook.CustomAttributesServiceNoEvent), new(*featurecustomattrs.ServiceNoEvent)),
	wire.Bind(new(featuresiwe.RateLimiter), new(*ratelimit.Limiter)),
	wire.Bind(new(identitysiwe.SIWEService), new(*featuresiwe.Service)),
	wire.Bind(new(identityservice.LoginIDIdentityProvider), new(*identityloginid.Provider)),
	wire.Bind(new(identityservice.OAuthIdentityProvider), new(*identityoauth.Provider)),
	wire.Bind(new(identityservice.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
	wire.Bind(new(identityservice.BiometricIdentityProvider), new(*identitybiometric.Provider)),
	wire.Bind(new(identityservice.PasskeyIdentityProvider), new(*identitypasskey.Provider)),
	wire.Bind(new(identityservice.SIWEIdentityProvider), new(*identitysiwe.Provider)),
	wire.Bind(new(user.AuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(identitypasskey.PasskeyService), new(*featurepasskey.Service)),
	wire.Bind(new(featurepasskey.TranslationService), new(*translation.Service)),
	wire.Bind(new(translation.StaticAssetResolver), new(*web.StaticAssetResolver)),
	wire.Bind(new(user.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(event.ResolverUserQueries), new(*user.Queries)),
	wire.Bind(new(identityloginid.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.EmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),
	wire.Bind(new(event.Database), new(*appdb.Handle)),
	wire.Bind(new(authenticatorservice.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
	wire.Bind(new(authenticatorservice.PasskeyAuthenticatorProvider), new(*authenticatorpasskey.Provider)),
	wire.Bind(new(authenticatorpasskey.PasskeyService), new(*featurepasskey.Service)),
	wire.Bind(new(authenticatorservice.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),
	wire.Bind(new(authenticatorservice.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
	wire.Bind(new(authenticatorservice.OTPCodeService), new(*otp.Service)),
	wire.Bind(new(otp.RateLimiter), new(*ratelimit.Limiter)),
	wire.Bind(new(authenticatorservice.RateLimiter), new(*ratelimit.Limiter)),
	wire.Bind(new(user.VerificationService), new(*verification.Service)),
	wire.Bind(new(user.StandardAttributesService), new(*featurestdattrs.ServiceNoEvent)),
	wire.Bind(new(featurestdattrs.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(featurestdattrs.UserQueries), new(*user.RawQueries)),
	wire.Bind(new(featurestdattrs.UserStore), new(*user.Store)),
	wire.Bind(new(featurestdattrs.ClaimStore), new(*verification.StorePQ)),
	wire.Bind(new(user.CustomAttributesService), new(*featurecustomattrs.ServiceNoEvent)),
	wire.Bind(new(featurecustomattrs.UserQueries), new(*user.RawQueries)),
	wire.Bind(new(featurecustomattrs.UserStore), new(*user.Store)),
	wire.Bind(new(user.Web3Service), new(*featureweb3.Service)),
	wire.Bind(new(libes.UserQueries), new(*user.Queries)),
)
