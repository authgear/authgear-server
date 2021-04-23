package config

type KubernetesConfig struct {
	// KubeConfig indicates the path to the `.kubeconfig` config file
	KubeConfig string `envconfig:"KUBERNETES_KUBECONFIG"`
	// AppNamespace indicates the namespace where the app's resources (e.g. ingress, cert) resides
	AppNamespace string `envconfig:"KUBERNETES_APP_NAMESPACE"`
}
