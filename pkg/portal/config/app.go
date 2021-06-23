package config

type SecretKeyAllowlist []string

type AppConfig struct {
	HostSuffix string              `envconfig:"HOST_SUFFIX"`
	IDPattern  string              `envconfig:"ID_PATTERN" default:"^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$"`
	Kubernetes AppKubernetesConfig `envconfig:"KUBERNETES"`

	// BuiltinResourceDirectory sets the directory for built-in resource files
	BuiltinResourceDirectory string `envconfig:"BUILTIN_RESOURCE_DIRECTORY" default:"resources/authgear"`
	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"CUSTOM_RESOURCE_DIRECTORY"`
	// MaxOwnedApps controls how many apps a user can own.
	MaxOwnedApps int `envconfig:"MAX_OWNED_APPS" default:"-1"`
	// SecretKeyAllowlist defines the allowlist of secret key.
	// If it is empty, then all keys are allowed.
	// Otherwise only allowed keys in the list are visible in API response.
	SecretKeyAllowlist SecretKeyAllowlist `envconfig:"SECRET_KEY_ALLOWLIST"`
	// DefaultPlan defines the default plan for apps during app creation
	DefaultPlan string `envconfig:"DEFAULT_PLAN"`
}

type AppKubernetesConfig struct {
	IngressTemplateFile string `envconfig:"INGRESS_TEMPLATE_FILE"`
}
