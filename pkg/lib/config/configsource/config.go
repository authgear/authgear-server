package configsource

type Type string

const (
	TypeLocalFS    Type = "local_fs"
	TypeKubernetes Type = "kubernetes"
)

var Types = []Type{
	TypeLocalFS,
	TypeKubernetes,
}

type Config struct {
	// Type sets the type of configuration source
	Type Type `envconfig:"TYPE" default:"local_fs"`

	// KubeConfig indicates the path to the `.kubeconfig` config file
	KubeConfig string `envconfig:"KUBECONFIG"`
	// KubeNamespace indicates the namespace where the app index & configs resides
	KubeNamespace string `envconfig:"KUBE_NAMESPACE"`
	// KubeIndexName indicates the name of app index ConfigMap
	KubeIndexName string `envconfig:"KUBE_INDEX_NAME" default:"app-index"`
	// KubeAppConfigPattern sets the RegEx that matches the app ID in ConfigMap/Secret name
	KubeAppConfigPattern string `envconfig:"KUBE_APP_CONFIG_PATTERN" default:"^app-data-(.+)$"`

	// Watch indicates whether the configuration source would watch for changes and reload automatically
	Watch bool `envconfig:"WATCH" default:"true"`
	// Directory sets the path to app configuration directory file for local FS sources
	Directory string `envconfig:"DIRECTORY" default:"."`
}
