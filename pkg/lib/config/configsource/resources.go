package configsource

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const (
	AuthgearYAML       = "authgear.yaml"
	AuthgearSecretYAML = "authgear.secrets.yaml"
)

var AppConfig = resource.RegisterResource(resource.SimpleFile{
	Name: AuthgearYAML,
	ParseFn: func(data []byte) (interface{}, error) {
		appConfig, err := config.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("cannot parse app config: %w", err)
		}
		return appConfig, nil
	},
})

var SecretConfig = resource.RegisterResource(secretConfigKind{})

type secretConfigKind struct{}

func (f secretConfigKind) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	data, err := resource.ReadFile(fs, AuthgearSecretYAML)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []resource.LayerFile{{Path: AuthgearSecretYAML, Data: data}}, nil
}

func (f secretConfigKind) MatchResource(path string) bool {
	return path == AuthgearSecretYAML
}

func (f secretConfigKind) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	var layerConfigs []*config.SecretConfig
	for _, layer := range layers {
		var layerConfig config.SecretConfig
		if err := yaml.Unmarshal(layer.Data, &layerConfig); err != nil {
			return nil, fmt.Errorf("malformed secret config: %w", err)
		}
		layerConfigs = append(layerConfigs, &layerConfig)
	}

	mergedConfig := (&config.SecretConfig{}).Overlay(layerConfigs...)
	mergedYAML, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return nil, err
	}

	return &resource.MergedFile{Data: mergedYAML}, nil
}

func (f secretConfigKind) Parse(merged *resource.MergedFile) (interface{}, error) {
	secretConfig, err := config.ParseSecret(merged.Data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse secret config: %w", err)
	}
	return secretConfig, nil
}
