package server

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Config struct {
	// ListenAddr sets the listen address of the portal server.
	PortalListenAddr         string `envconfig:"PORTAL_LISTEN_ADDR" default:"0.0.0.0:3003"`
	PortalInternalListenAddr string `envconfig:"PORTAL_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13003"`
	// SiteadminListenAddr sets the listen address of the Site Admin API server.
	SiteadminListenAddr         string `envconfig:"SITEADMIN_LISTEN_ADDR" default:"0.0.0.0:3005"`
	SiteadminInternalListenAddr string `envconfig:"SITEADMIN_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13005"`
	// ConfigSource configures the source of app configurations
	ConfigSource *configsource.Config `envconfig:"CONFIG_SOURCE"`
	// Authgear configures Authgear acting as authentication server for the portal.
	Authgear portalconfig.AuthgearConfig `envconfig:"AUTHGEAR"`
	// SiteadminAuthgear configures Authgear for the Site Admin API.
	// Allows the siteadmin server to authenticate against a different Authgear app than the portal.
	SiteadminAuthgear portalconfig.AuthgearConfig `envconfig:"SITEADMIN_AUTHGEAR"`
	// AdminAPI configures how portal interacts with Authgear Admin API.
	AdminAPI portalconfig.AdminAPIConfig `envconfig:"ADMIN_API"`
	// App configures the managed apps.
	App portalconfig.AppConfig `envconfig:"APP"`
	// SMTP configures SMTP.
	SMTP portalconfig.SMTPConfig `envconfig:"SMTP"`
	// Mail configures email settings.
	Mail portalconfig.MailConfig `envconfig:"MAIL"`

	// PORTAL_BUILTIN_RESOURCE_DIRECTORY is deprecated. It has no effect anymore.

	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"PORTAL_CUSTOM_RESOURCE_DIRECTORY"`

	// DomainImplementation indicates the domain implementation, only none or kubernetes are supported
	// if it sets to kubernetes, kubernetes resources will be created based on
	// APP_KUBERNETES_INGRESS_TEMPLATE_FILE when creating domains
	DomainImplementation portalconfig.DomainImplementationType `envconfig:"DOMAIN_IMPLEMENTATION"`

	// Kubernetes set the kubernetes related config if the portal is deployed in k8s
	// One of the purpose is for creating ingress when creating new domain
	Kubernetes portalconfig.KubernetesConfig `envconfig:"KUBERNETES"`

	// Search sets search related config.
	Search portalconfig.SearchConfig `envconfig:"SEARCH"`

	// AuditLog sets audit log related config.
	AuditLog portalconfig.AuditLogConfig `envconfig:"AUDIT_LOG"`

	// Analytic sets analytic dashboard related config.
	Analytic config.AnalyticConfig `envconfig:"ANALYTIC"`

	Stripe portalconfig.StripeConfig `envconfig:"STRIPE"`

	Osano portalconfig.OsanoConfig `envconfig:"OSANO"`

	GoogleTagManager portalconfig.GoogleTagManagerConfig `envconfig:"GTM"`

	PortalFrontendSentry portalconfig.PortalFrontendSentryConfig `envconfig:"PORTAL_FRONTEND_SENTRY"`
	PortalFeatures       portalconfig.PortalFeaturesConfig       `envconfig:"PORTAL_FEATURES"`

	*config.EnvironmentConfig
}

type LoadConfigOptions struct {
	ServePortal    bool
	ServeSiteadmin bool
}

func LoadConfigFromEnv(opts LoadConfigOptions) (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	err = config.Validate(opts)
	if err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}

	return config, nil
}

func (c *Config) Validate(opts LoadConfigOptions) error {
	ctx := &validation.Context{}

	sourceTypes := make([]string, len(configsource.Types))
	ok := false
	for i, t := range configsource.Types {
		if t == c.ConfigSource.Type {
			ok = true
			break
		}
		sourceTypes[i] = string(t)
	}
	if !ok {
		ctx.Child("CONFIG_SOURCE_TYPE").EmitErrorMessage(
			"invalid configuration source type; available: " + strings.Join(sourceTypes, ", "),
		)
	}

	if c.GlobalDatabase.DatabaseURL == "" {
		ctx.Child("DATABASE_URL").EmitErrorMessage("missing database URL")
	}

	if opts.ServePortal {
		if c.Authgear.ClientID == "" {
			ctx.Child("AUTHGEAR_CLIENT_ID").EmitErrorMessage("missing authgear client ID")
		}
		if c.Authgear.Endpoint == "" {
			ctx.Child("AUTHGEAR_ENDPOINT").EmitErrorMessage("missing authgear endpoint")
		}
	}

	if opts.ServeSiteadmin {
		if c.SiteadminAuthgear.AppID == "" {
			ctx.Child("SITEADMIN_AUTHGEAR_APP_ID").EmitErrorMessage("missing siteadmin authgear app ID")
		}
		if c.SiteadminAuthgear.Endpoint == "" {
			ctx.Child("SITEADMIN_AUTHGEAR_ENDPOINT").EmitErrorMessage("missing siteadmin authgear endpoint")
		}
	}

	// Stripe config is optional

	return ctx.Error("invalid server configuration")
}
