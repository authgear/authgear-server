package kubeutil

import (
	"errors"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func MakeKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath == "" {
		kubeConfig, err := rest.InClusterConfig()
		if errors.Is(err, rest.ErrNotInCluster) {
			kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
			kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		}
		if err != nil {
			return nil, err
		}
		return kubeConfig, nil
	}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}
