package configsource

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const (
	AuthgearYAML        = "authgear.yaml"
	AuthgearSecretYAML  = "authgear.secrets.yaml"
	AuthgearFeatureYAML = "authgear.features.yaml"
)

var ErrEffectiveSecretConfig = apierrors.NewForbidden("cannot view effective secret config")

type ViewWithConfig interface {
	resource.View
	SecretKeyAllowlist() []string
	AppFeatureConfig() *config.FeatureConfig
}

type AuthgearYAMLDescriptor struct{}

var _ resource.Descriptor = AuthgearYAMLDescriptor{}

func (d AuthgearYAMLDescriptor) MatchResource(path string) (*resource.Match, bool) {
	if path == AuthgearYAML {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d AuthgearYAMLDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	location := resource.Location{
		Fs:   fs,
		Path: AuthgearYAML,
	}
	_, err := resource.ReadLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []resource.Location{location}, nil
}

func (d AuthgearYAMLDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	app := func() (interface{}, error) {
		var target *resource.ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp {
				s := resrc
				target = &s
			}
		}
		if target == nil {
			return nil, resource.ErrResourceNotFound
		}

		return target.Data, nil
	}

	effective := func() (interface{}, error) {
		bytes, err := app()
		if err != nil {
			return nil, err
		}

		appConfig, err := config.Parse(bytes.([]byte))
		if err != nil {
			return nil, fmt.Errorf("cannot parse app config: %w", err)
		}
		return appConfig, nil
	}

	switch rawView.(type) {
	case resource.AppFileView:
		return app()
	case resource.EffectiveFileView:
		return app()
	case resource.EffectiveResourceView:
		return effective()
	case resource.ValidateResourceView:
		return effective()
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d AuthgearYAMLDescriptor) UpdateResource(resrc *resource.ResourceFile, data []byte, rawView resource.View) (*resource.ResourceFile, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot delete '%v'", AuthgearYAML)
	}

	view, ok := rawView.(resource.AppFileView)
	if !ok {
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}

	viewWithConfig, ok := view.(interface{}).(ViewWithConfig)
	if !ok {
		panic("resource: missing config in the app file view")
	}

	fc := viewWithConfig.AppFeatureConfig()

	original, err := config.Parse(resrc.Data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse original app config %w", err)
	}

	incoming, err := config.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse incoming app config: %w", err)
	}

	err = d.validate(original, incoming, fc)
	if err != nil {
		return nil, err
	}

	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func (d AuthgearYAMLDescriptor) validate(original *config.AppConfig, incoming *config.AppConfig, fc *config.FeatureConfig) error {
	// check when adding new oauth clients
	if len(original.OAuth.Clients) < len(incoming.OAuth.Clients) {
		if len(incoming.OAuth.Clients) > *fc.OAuth.Client.Maximum {
			return fmt.Errorf("exceed the maximum number of oauth clients, actual: %d, expected: %d",
				len(incoming.OAuth.Clients),
				*fc.OAuth.Client.Maximum,
			)
		}
	}

	if len(original.Identity.OAuth.Providers) < len(incoming.Identity.OAuth.Providers) {
		if len(incoming.Identity.OAuth.Providers) > *fc.Identity.OAuth.MaximumProviders {
			return fmt.Errorf("exceed the maximum number of sso providers, actual: %d, expected: %d",
				len(incoming.Identity.OAuth.Providers),
				*fc.Identity.OAuth.MaximumProviders,
			)
		}
	}

	return nil
}

var AppConfig = resource.RegisterResource(AuthgearYAMLDescriptor{})

type AuthgearSecretYAMLDescriptor struct{}

var _ resource.Descriptor = AuthgearSecretYAMLDescriptor{}

func (d AuthgearSecretYAMLDescriptor) MatchResource(path string) (*resource.Match, bool) {
	if path == AuthgearSecretYAML {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d AuthgearSecretYAMLDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	location := resource.Location{
		Fs:   fs,
		Path: AuthgearSecretYAML,
	}
	_, err := resource.ReadLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []resource.Location{location}, nil
}

func (d AuthgearSecretYAMLDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch view := rawView.(type) {
	case resource.AppFileView:
		return d.viewAppFile(resources, view)
	case resource.EffectiveFileView:
		return nil, ErrEffectiveSecretConfig
	case resource.EffectiveResourceView:
		return d.viewEffectiveResource(resources)
	case resource.ValidateResourceView:
		return d.viewEffectiveResource(resources)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d AuthgearSecretYAMLDescriptor) viewAppFile(resources []resource.ResourceFile, view resource.AppFileView) (interface{}, error) {
	viewWithConfig, ok := view.(interface{}).(ViewWithConfig)
	if !ok {
		panic("resource: missing config in the app file view")
	}
	allowlist := viewWithConfig.SecretKeyAllowlist()

	var target *resource.ResourceFile
	for _, resrc := range resources {
		if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp {
			s := resrc
			target = &s
		}
	}

	if target == nil {
		return nil, resource.ErrResourceNotFound
	}

	var cfg config.SecretConfig
	if err := yaml.Unmarshal(target.Data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed secret config: %w", err)
	}

	if len(allowlist) > 0 {
		allowmap := make(map[config.SecretKey]struct{})
		for _, key := range allowlist {
			allowmap[config.SecretKey(key)] = struct{}{}
		}

		var secrets []config.SecretItem
		for _, secretItem := range cfg.Secrets {
			_, allowed := allowmap[secretItem.Key]
			if allowed {
				secrets = append(secrets, secretItem)
			}
		}
		cfg.Secrets = secrets
	}

	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret config: %w", err)
	}

	return bytes, nil
}

func (d AuthgearSecretYAMLDescriptor) viewEffectiveResource(resources []resource.ResourceFile) (interface{}, error) {
	var cfgs []*config.SecretConfig
	for _, layer := range resources {
		var cfg config.SecretConfig
		if err := yaml.Unmarshal(layer.Data, &cfg); err != nil {
			return nil, fmt.Errorf("malformed secret config: %w", err)
		}
		cfgs = append(cfgs, &cfg)
	}

	mergedConfig := (&config.SecretConfig{}).Overlay(cfgs...)
	mergedYAML, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return nil, err
	}

	secretConfig, err := config.ParseSecret(mergedYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot parse secret config: %w", err)
	}
	return secretConfig, nil
}

func (d AuthgearSecretYAMLDescriptor) UpdateResource(resrc *resource.ResourceFile, data []byte, rawView resource.View) (*resource.ResourceFile, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot delete '%v'", AuthgearSecretYAML)
	}

	switch view := rawView.(type) {
	case resource.AppFileView:
		var original config.SecretConfig
		err := yaml.Unmarshal(resrc.Data, &original)
		if err != nil {
			return nil, fmt.Errorf("failed to parse original secret config: %w", err)
		}

		var incoming config.SecretConfig
		err = yaml.Unmarshal(data, &incoming)
		if err != nil {
			return nil, fmt.Errorf("failed to parse incoming secret config: %w", err)
		}

		viewWithConfig, ok := view.(interface{}).(ViewWithConfig)
		if !ok {
			panic("resource: missing config in the app file view")
		}
		allowlist := viewWithConfig.SecretKeyAllowlist()

		// When allowlist is non-empty:
		// For example, suppose original has "a", "b", "c" and the allowlist is "a".
		// Then original should keep "b" and "c" only.
		//
		// When allowlist is empty:
		// Then original should be ignored.
		var mergedConfig *config.SecretConfig
		if len(allowlist) > 0 {
			allowmap := make(map[config.SecretKey]struct{})
			for _, key := range allowlist {
				allowmap[config.SecretKey(key)] = struct{}{}
			}

			for _, secretItem := range incoming.Secrets {
				_, allowed := allowmap[secretItem.Key]
				if !allowed {
					return nil, fmt.Errorf("'%s' in secret config is not allowed", secretItem.Key)
				}
			}

			var originalSecrets []config.SecretItem
			for _, secretItem := range original.Secrets {
				_, allowed := allowmap[secretItem.Key]
				if !allowed {
					originalSecrets = append(originalSecrets, secretItem)
				}
			}

			mergedConfig = (&config.SecretConfig{}).Overlay(&config.SecretConfig{
				Secrets: originalSecrets,
			}, &incoming)
		} else {
			mergedConfig = &incoming
		}

		mergedYAML, err := yaml.Marshal(&mergedConfig)
		if err != nil {
			return nil, err
		}

		newResrc := *resrc
		newResrc.Data = mergedYAML
		return &newResrc, nil
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

var SecretConfig = resource.RegisterResource(AuthgearSecretYAMLDescriptor{})

type AuthgearFeatureYAMLDescriptor struct{}

var _ resource.Descriptor = AuthgearFeatureYAMLDescriptor{}

func (d AuthgearFeatureYAMLDescriptor) MatchResource(path string) (*resource.Match, bool) {
	if path == AuthgearFeatureYAML {
		return &resource.Match{}, true
	}
	return nil, false
}

func (d AuthgearFeatureYAMLDescriptor) FindResources(fs resource.Fs) ([]resource.Location, error) {
	location := resource.Location{
		Fs:   fs,
		Path: AuthgearFeatureYAML,
	}
	_, err := resource.ReadLocation(location)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []resource.Location{location}, nil
}

func (d AuthgearFeatureYAMLDescriptor) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	app := func() (interface{}, error) {
		var target *resource.ResourceFile
		for _, resrc := range resources {
			if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp {
				s := resrc
				target = &s
			}
		}
		if target == nil {
			return nil, resource.ErrResourceNotFound
		}

		return target.Data, nil
	}

	effective := func() (interface{}, error) {
		bytes, err := app()
		if err != nil {
			return nil, err
		}

		featureConfig, err := config.ParseFeatureConfig(bytes.([]byte))
		if err != nil {
			return nil, fmt.Errorf("cannot parse feature config: %w", err)
		}
		return featureConfig, nil
	}

	switch rawView.(type) {
	case resource.AppFileView:
		return app()
	case resource.EffectiveFileView:
		return app()
	case resource.EffectiveResourceView:
		return effective()
	case resource.ValidateResourceView:
		return effective()
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d AuthgearFeatureYAMLDescriptor) UpdateResource(resrc *resource.ResourceFile, data []byte, _ resource.View) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("cannot update '%v'", AuthgearFeatureYAML)
}

var FeatureConfig = resource.RegisterResource(AuthgearFeatureYAMLDescriptor{})
