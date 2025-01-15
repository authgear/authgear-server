package config

type AppConfig struct {
	HostSuffix string              `envconfig:"HOST_SUFFIX" default:".localhost:3002"`
	IDPattern  string              `envconfig:"ID_PATTERN" default:"^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$"`
	Kubernetes AppKubernetesConfig `envconfig:"KUBERNETES"`

	// BUILTIN_RESOURCE_DIRECTORY is deprecated. It has no effect anymore.

	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"CUSTOM_RESOURCE_DIRECTORY"`
	// MaxOwnedApps controls how many apps a user can own.
	MaxOwnedApps int `envconfig:"MAX_OWNED_APPS" default:"-1"`
	// DefaultPlan defines the default plan for apps during app creation
	DefaultPlan string `envconfig:"DEFAULT_PLAN"`
}

type AppKubernetesConfig struct {
	IngressTemplateFile string `envconfig:"INGRESS_TEMPLATE_FILE"`
}
