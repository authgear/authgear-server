package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ProvideRequestContext(r *http.Request) context.Context { return r.Context() }

func ProvideConfigSource(ctrl *configsource.Controller) *configsource.ConfigSource {
	return ctrl.GetConfigSource()
}

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"AuthgearConfig",
		"AdminAPIConfig",
		"AppConfig",
		"DatabaseConfig",
		"SMTPConfig",
		"MailConfig",
		"KubernetesConfig",
		"SentryHub",
		"LoggerFactory",
		"Database",
		"ConfigSourceController",
		"Resources",
		"SecretKeyAllowlist",
	),
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
		"Request",
	),
	ProvideRequestContext,
	ProvideConfigSource,
	ProvideAppBaseResources,
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Value(template.DefaultLanguageTag(intl.DefaultLanguage)),
	wire.Value(template.SupportedLanguageTags([]string{intl.DefaultLanguage})),
)
