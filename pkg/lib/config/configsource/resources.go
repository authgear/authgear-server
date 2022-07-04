package configsource

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/readcloserthunk"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	AuthgearYAML        = "authgear.yaml"
	AuthgearSecretYAML  = "authgear.secrets.yaml"
	AuthgearFeatureYAML = "authgear.features.yaml"
)

var ErrEffectiveSecretConfig = apierrors.NewForbidden("cannot view effective secret config")

type contextKeyFeatureConfigType string

const ContextKeyFeatureConfig = contextKeyFeatureConfigType(AuthgearFeatureYAML)

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
	_, err := resource.StatLocation(location)
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

		b, err := readcloserthunk.Performance_Bytes(target.ReadCloserThunk)
		if err != nil {
			return nil, err
		}

		return b, nil
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

func (d AuthgearYAMLDescriptor) UpdateResource(ctx context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot delete '%v'", AuthgearYAML)
	}

	fc, ok := ctx.Value(ContextKeyFeatureConfig).(*config.FeatureConfig)
	if !ok || fc == nil {
		return nil, fmt.Errorf("missing feature config in context")
	}

	b, err := readcloserthunk.Performance_Bytes(resrc.ReadCloserThunk)
	if err != nil {
		return nil, err
	}

	original, err := config.Parse(b)
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
		Location:        resrc.Location,
		ReadCloserThunk: readcloserthunk.Reader(bytes.NewReader(data)),
	}, nil
}

func (d AuthgearYAMLDescriptor) validate(original *config.AppConfig, incoming *config.AppConfig, fc *config.FeatureConfig) error {
	validationCtx := &validation.Context{}

	// Enforce feature config.
	if len(original.OAuth.Clients) < len(incoming.OAuth.Clients) {
		if len(incoming.OAuth.Clients) > *fc.OAuth.Client.Maximum {
			validationCtx.Child(
				"oauth",
				"clients",
			).EmitErrorMessage(
				fmt.Sprintf("exceed the maximum number of oauth clients, actual: %d, expected: %d",
					len(incoming.OAuth.Clients),
					*fc.OAuth.Client.Maximum,
				),
			)
		}
	}
	if len(original.Identity.OAuth.Providers) < len(incoming.Identity.OAuth.Providers) {
		if len(incoming.Identity.OAuth.Providers) > *fc.Identity.OAuth.MaximumProviders {
			validationCtx.Child(
				"identity",
				"oauth",
				"providers",
			).EmitErrorMessage(
				fmt.Sprintf("exceed the maximum number of sso providers, actual: %d, expected: %d",
					len(incoming.Identity.OAuth.Providers),
					*fc.Identity.OAuth.MaximumProviders,
				),
			)
		}
	}
	if len(original.Hook.BlockingHandlers) < len(incoming.Hook.BlockingHandlers) {
		if len(incoming.Hook.BlockingHandlers) > *fc.Hook.BlockingHandler.Maximum {
			validationCtx.Child(
				"hook",
				"blocking_handlers",
			).EmitErrorMessage(
				fmt.Sprintf("exceed the maximum number of blocking handlers, actual: %d, expected: %d",
					len(incoming.Hook.BlockingHandlers),
					*fc.Hook.BlockingHandler.Maximum,
				),
			)
		}
	}
	if len(original.Hook.NonBlockingHandlers) < len(incoming.Hook.NonBlockingHandlers) {
		if len(incoming.Hook.NonBlockingHandlers) > *fc.Hook.NonBlockingHandler.Maximum {
			validationCtx.Child(
				"hook",
				"non_blocking_handlers",
			).EmitErrorMessage(
				fmt.Sprintf("exceed the maximum number of non blocking handlers, actual: %d, expected: %d",
					len(incoming.Hook.NonBlockingHandlers),
					*fc.Hook.NonBlockingHandler.Maximum,
				),
			)
		}
	}

	// Check custom attributes not removed nor edited.
	for _, originalCustomAttr := range original.UserProfile.CustomAttributes.Attributes {
		found := false
		for i, incomingCustomAttr := range incoming.UserProfile.CustomAttributes.Attributes {
			if originalCustomAttr.ID == incomingCustomAttr.ID {
				found = true
				if originalCustomAttr.Type != incomingCustomAttr.Type {
					validationCtx.Child(
						"user_profile",
						"custom_attributes",
						"attributes",
						strconv.Itoa(i),
					).EmitErrorMessage(
						fmt.Sprintf("custom attribute of id '%v' has type changed; original: %v, incoming: %v",
							originalCustomAttr.ID,
							originalCustomAttr.Type,
							incomingCustomAttr.Type,
						),
					)
				}
			}
		}
		if !found {
			validationCtx.Child(
				"user_profile",
				"custom_attributes",
				"attributes",
			).EmitErrorMessage(
				fmt.Sprintf("custom attribute of id '%v' was deleted", originalCustomAttr.ID),
			)
		}
	}

	return validationCtx.Error(fmt.Sprintf("invalid %v", AuthgearYAML))
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
	_, err := resource.StatLocation(location)
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
	b, err := readcloserthunk.Performance_Bytes(target.ReadCloserThunk)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("malformed secret config: %w", err)
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
		b, err := readcloserthunk.Performance_Bytes(layer.ReadCloserThunk)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(b, &cfg); err != nil {
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

func (d AuthgearSecretYAMLDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot delete '%v'", AuthgearSecretYAML)
	}

	var original config.SecretConfig
	b, err := readcloserthunk.Performance_Bytes(resrc.ReadCloserThunk)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &original)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original secret config: %w", err)
	}

	var incoming config.SecretConfig
	err = yaml.Unmarshal(data, &incoming)
	if err != nil {
		return nil, fmt.Errorf("failed to parse incoming secret config: %w", err)
	}

	updatedConfig := original.UpdateWith(&incoming)
	updatedYAML, err := yaml.Marshal(updatedConfig)
	if err != nil {
		return nil, err
	}

	newResrc := *resrc
	newResrc.ReadCloserThunk = readcloserthunk.Reader(bytes.NewReader(updatedYAML))
	return &newResrc, nil
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
	_, err := resource.StatLocation(location)
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

		b, err := readcloserthunk.Performance_Bytes(target.ReadCloserThunk)
		if err != nil {
			return nil, err
		}

		return b, nil
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

func (d AuthgearFeatureYAMLDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("cannot update '%v'", AuthgearFeatureYAML)
}

var FeatureConfig = resource.RegisterResource(AuthgearFeatureYAMLDescriptor{})
