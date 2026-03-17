package configsource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/secrets"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate go tool mockgen -source=resources.go -destination=resources_mock_test.go -package configsource

const (
	AuthgearYAML        = "authgear.yaml"
	AuthgearSecretYAML  = "authgear.secrets.yaml"
	AuthgearFeatureYAML = "authgear.features.yaml"
)

type DomainService interface {
	ListDomains(ctx context.Context, appID string) ([]*apimodel.Domain, error)
}

var ErrEffectiveSecretConfig = apierrors.NewForbidden("cannot view effective secret config")

type contextKeyFeatureConfigType string

const ContextKeyFeatureConfig = contextKeyFeatureConfigType(AuthgearFeatureYAML)

type contextKeyAppHostSuffixes string

var ContextKeyAppHostSuffixes = contextKeyAppHostSuffixes("APP_HOST_SUFFIXES")

type domainServiceContextKeyType struct{}

var ContextKeyDomainService = domainServiceContextKeyType{}

type clockContextKeyType struct{}

var ContextKeyClock = clockContextKeyType{}

type samlEntityIDKeyType struct{}

var ContextKeySAMLEntityID = samlEntityIDKeyType{}

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

func (d AuthgearYAMLDescriptor) ViewResources(ctx context.Context, resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
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

		appConfig, err := config.Parse(ctx, bytes.([]byte))
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

	appHostSuffixes, ok := ctx.Value(ContextKeyAppHostSuffixes).(*config.AppHostSuffixes)
	if !ok {
		return nil, fmt.Errorf("missing app host suffixes in context")
	}

	domainService, ok := ctx.Value(ContextKeyDomainService).(DomainService)
	if !ok || domainService == nil {
		return nil, fmt.Errorf("missing domain service in context")
	}

	original, err := config.Parse(ctx, resrc.Data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse original app config %w", err)
	}

	incoming, err := config.Parse(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse incoming app config: %w", err)
	}

	err = d.validate(ctx, original, incoming, fc, *appHostSuffixes, domainService)
	if err != nil {
		return nil, err
	}

	return &resource.ResourceFile{
		Location: resrc.Location,
		Data:     data,
	}, nil
}

func (d AuthgearYAMLDescriptor) validate(ctx context.Context, original *config.AppConfig, incoming *config.AppConfig, fc *config.FeatureConfig, appHostSuffixes []string, domainService DomainService) error {
	validationCtx := &validation.Context{}

	d.validateCustomAttributes(validationCtx, original, incoming)
	d.validateFeatureConfig(validationCtx, incoming, original, fc)
	err := d.validatePublicOrigin(ctx, validationCtx, incoming, original, appHostSuffixes, domainService)
	if err != nil {
		return err
	}
	d.validateOAuthClients(validationCtx, incoming, original)

	return validationCtx.Error(fmt.Sprintf("invalid %v", AuthgearYAML))
}

// Check custom attributes not removed nor edited.
func (d AuthgearYAMLDescriptor) validateCustomAttributes(validationCtx *validation.Context, original *config.AppConfig, incoming *config.AppConfig) {
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
}

// Enforce feature config.
func (d AuthgearYAMLDescriptor) validateFeatureConfig(validationCtx *validation.Context, incoming *config.AppConfig, original *config.AppConfig, fc *config.FeatureConfig) {
	featureConfigErr := func() error {
		incomingFCError := d.validateBasedOnFeatureConfig(incoming, fc)
		incomingAggregatedError, ok := incomingFCError.(*validation.AggregatedError)
		if incomingFCError == nil || !ok {
			return incomingFCError
		}
		// https://github.com/authgear/authgear-server/commit/888e57b4b6fa9de7cd5786111cdc5cc244a85ac0
		// If the original config has some feature config error, we allow the user
		// to save the config without correcting them. This is for the case that
		// the app is downgraded from a higher plan.
		originalFCError := d.validateBasedOnFeatureConfig(original, fc)
		originalAggregatedError, ok := originalFCError.(*validation.AggregatedError)
		if originalFCError == nil || !ok {
			return incomingFCError
		}

		aggregatedError := incomingAggregatedError.Subtract(originalAggregatedError)
		return aggregatedError
	}()

	validationCtx.AddError(featureConfigErr)

	// fraud_protection is guarded by is_modifiable.
	// We allow the save when:
	//   - incoming == original (unchanged), OR
	//   - incoming == default (nil / omitted from the YAML, meaning "don't touch it")
	// This handles plan downgrades: a config customised while is_modifiable=true
	// can still be saved unchanged after the plan is downgraded.
	if !*fc.FraudProtection.IsModifiable {
		defaultFP := &config.FraudProtectionConfig{}
		config.SetFieldDefaults(defaultFP)
		incomingFPJSON, _ := json.Marshal(incoming.FraudProtection)
		originalFPJSON, _ := json.Marshal(original.FraudProtection)
		defaultFPJSON, _ := json.Marshal(defaultFP)
		if string(incomingFPJSON) != string(originalFPJSON) && string(incomingFPJSON) != string(defaultFPJSON) {
			validationCtx.Child("fraud_protection").EmitErrorMessage("fraud_protection config is not modifiable")
		}
	}
}

// Check public origin.
func (d AuthgearYAMLDescriptor) validatePublicOrigin(ctx context.Context, validationCtx *validation.Context, incoming *config.AppConfig, original *config.AppConfig, appHostSuffixes []string, domainService DomainService) error {
	if incoming.HTTP.PublicOrigin != original.HTTP.PublicOrigin {
		validOrigin := false

		incomingUrl, err := url.Parse(incoming.HTTP.PublicOrigin)
		if err != nil {
			return err
		}

		for _, appHostSuffix := range appHostSuffixes {
			appHost := string(incoming.ID) + appHostSuffix
			if incomingUrl.Host == appHost {
				validOrigin = true
				break
			}
		}

		availableDomains, err := domainService.ListDomains(ctx, string(incoming.ID))
		if err != nil {
			return err
		}

		for _, domain := range availableDomains {
			if incomingUrl.Host == domain.Domain {
				validOrigin = true
				break
			}
		}

		if !validOrigin {
			validationCtx.Child(
				"http",
				"public_origin",
			).EmitErrorMessage(
				fmt.Sprintf("public origin is not allowed"),
			)
		}
	}

	return nil
}

func (d AuthgearYAMLDescriptor) validateOAuthClients(validationCtx *validation.Context, incoming *config.AppConfig, original *config.AppConfig) {
	incomingClientIds := map[string]struct{}{}
	for _, incomingClient := range incoming.OAuth.Clients {
		incomingClientIds[incomingClient.ClientID] = struct{}{}
	}
	origClientIds := map[string]struct{}{}
	for _, origClient := range original.OAuth.Clients {
		origClientIds[origClient.ClientID] = struct{}{}
	}
	var addedClientIds []string
	for incomingClientId := range incomingClientIds {
		if _, origHasIncoming := origClientIds[incomingClientId]; !origHasIncoming {
			addedClientIds = append(addedClientIds, incomingClientId)
		}
	}
	var removedClientIds []string
	for origClientId := range origClientIds {
		if _, incomingHasOrig := incomingClientIds[origClientId]; !incomingHasOrig {
			removedClientIds = append(removedClientIds, origClientId)
		}
	}
	// Ref DEV-1146 Disallow Changing Client ID
	// - authgear portal will not add and remove client at the same operation
	// - if there is both added and removed clients, user is probably modifying client id manually via api
	if len(addedClientIds) > 0 && len(removedClientIds) > 0 {
		validationCtx.Child("oauth", "clients").EmitErrorMessage("client ids cannot be changed")
	}

	// Validate the final clients length <= 50
	const maxFinalOAuthClients = 50
	if len(incoming.OAuth.Clients) > maxFinalOAuthClients {
		validationCtx.Child("oauth", "clients").EmitErrorMessage(
			fmt.Sprintf("exceed the maximum number of oauth clients, actual: %d, expected: %d", len(incoming.OAuth.Clients), maxFinalOAuthClients),
		)
	}
}

func (d AuthgearYAMLDescriptor) validateBasedOnFeatureConfig(appConfig *config.AppConfig, fc *config.FeatureConfig) error {
	validationCtx := &validation.Context{}

	if len(appConfig.OAuth.Clients) > *fc.OAuth.Client.Maximum {
		validationCtx.Child(
			"oauth",
			"clients",
		).EmitErrorMessage(
			fmt.Sprintf("exceed the maximum number of oauth clients, actual: %d, expected: %d",
				len(appConfig.OAuth.Clients),
				*fc.OAuth.Client.Maximum,
			),
		)
	}

	if len(appConfig.Identity.OAuth.Providers) > *fc.Identity.OAuth.MaximumProviders {
		validationCtx.Child(
			"identity",
			"oauth",
			"providers",
		).EmitErrorMessage(
			fmt.Sprintf("exceed the maximum number of sso providers, actual: %d, expected: %d",
				len(appConfig.Identity.OAuth.Providers),
				*fc.Identity.OAuth.MaximumProviders,
			),
		)
	}

	if len(appConfig.Hook.BlockingHandlers) > *fc.Hook.BlockingHandler.Maximum {
		validationCtx.Child(
			"hook",
			"blocking_handlers",
		).EmitErrorMessage(
			fmt.Sprintf("exceed the maximum number of blocking handlers, actual: %d, expected: %d",
				len(appConfig.Hook.BlockingHandlers),
				*fc.Hook.BlockingHandler.Maximum,
			),
		)
	}

	if len(appConfig.Hook.NonBlockingHandlers) > *fc.Hook.NonBlockingHandler.Maximum {
		validationCtx.Child(
			"hook",
			"non_blocking_handlers",
		).EmitErrorMessage(
			fmt.Sprintf("exceed the maximum number of non blocking handlers, actual: %d, expected: %d",
				len(appConfig.Hook.NonBlockingHandlers),
				*fc.Hook.NonBlockingHandler.Maximum,
			),
		)
	}

	for _, identity := range appConfig.Authentication.Identities {
		if identity == model.IdentityTypeBiometric && *fc.Identity.Biometric.Disabled {
			validationCtx.Child(
				"authentication",
				"identities",
			).EmitErrorMessage("enabling biometric authentication is not supported")
		}
	}

	// Password policy
	if *fc.Authenticator.Password.Policy.MinimumGuessableLevel.Disabled {
		if appConfig.Authenticator.Password.Policy.MinimumGuessableLevel != 0 {
			validationCtx.Child(
				"authenticator",
				"password",
				"policy",
				"minimum_guessable_level",
			).EmitErrorMessage("minimum_guessable_level is disallowed")
		}
	}
	if *fc.Authenticator.Password.Policy.ExcludedKeywords.Disabled {
		if len(appConfig.Authenticator.Password.Policy.ExcludedKeywords) > 0 {
			validationCtx.Child(
				"authenticator",
				"password",
				"policy",
				"excluded_keywords",
			).EmitErrorMessage("excluded_keywords is disallowed")
		}
	}
	if *fc.Authenticator.Password.Policy.History.Disabled {
		if appConfig.Authenticator.Password.Policy.IsEnabled() {
			validationCtx.Child(
				"authenticator",
				"password",
				"policy",
			).EmitErrorMessage("password history is disallowed")
		}
	}
	if !*fc.OAuth.Client.CustomUIEnabled {
		for i, oauthClient := range appConfig.OAuth.Clients {
			if oauthClient.CustomUIURI != "" {
				validationCtx.Child(
					"oauth",
					"clients",
					strconv.Itoa(i),
				).EmitErrorMessage("custom ui is disallowed")
			}
		}
	}

	return validationCtx.Error("features are limited by feature config")
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

func (d AuthgearSecretYAMLDescriptor) ViewResources(ctx context.Context, resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch view := rawView.(type) {
	case resource.AppFileView:
		return d.viewAppFile(resources, view)
	case resource.EffectiveFileView:
		return nil, ErrEffectiveSecretConfig
	case resource.EffectiveResourceView:
		return d.viewEffectiveResource(ctx, resources)
	case resource.ValidateResourceView:
		return d.viewEffectiveResource(ctx, resources)
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
	if err := yaml.Unmarshal(target.Data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed secret config: %w", err)
	}

	for idx, _ := range cfg.Secrets {
		cfg.Secrets[idx].FsLevel = resource.FsLevelApp
	}

	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret config: %w", err)
	}

	return bytes, nil
}

func (d AuthgearSecretYAMLDescriptor) viewEffectiveResource(ctx context.Context, resources []resource.ResourceFile) (interface{}, error) {
	var cfgs []*config.SecretConfig
	for _, layer := range resources {
		var cfg config.SecretConfig
		if err := yaml.Unmarshal(layer.Data, &cfg); err != nil {
			return nil, fmt.Errorf("malformed secret config: %w", err)
		}
		for idx, _ := range cfg.Secrets {
			cfg.Secrets[idx].FsLevel = layer.Location.Fs.GetFsLevel()
		}
		cfgs = append(cfgs, &cfg)
	}

	mergedConfig := (&config.SecretConfig{}).Overlay(cfgs...)
	mergedYAML, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return nil, err
	}

	mergedConfig, err = config.ParseSecretData(ctx, *mergedConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot parse secret config data: %w", err)
	}

	// Validate the merged config by parsing it
	// Do not use the return value because it unset FsLevel
	_, err = config.ParseSecret(ctx, mergedYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot parse secret config: %w", err)
	}
	return mergedConfig, nil
}

func (d AuthgearSecretYAMLDescriptor) UpdateResource(ctx context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot delete '%v'", AuthgearSecretYAML)
	}

	fc, ok := ctx.Value(ContextKeyFeatureConfig).(*config.FeatureConfig)
	if !ok || fc == nil {
		return nil, fmt.Errorf("missing feature config in context")
	}

	var original *config.SecretConfig
	original, err := config.ParseSecret(ctx, resrc.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original secret config: %w", err)
	}

	updateInstruction, err := ParseAuthgearSecretsYAMLUpdateInstructions(data)
	if err != nil {
		return nil, err
	}

	c, ok := ctx.Value(ContextKeyClock).(clock.Clock)
	if !ok || c == nil {
		return nil, fmt.Errorf("missing clock in context")
	}

	commonName := ctx.Value(ContextKeySAMLEntityID).(string)

	updateInstructionContext := &config.SecretConfigUpdateInstructionContext{
		Clock: c,
		// The key generated for client secret doesn't have use usage key
		// Since the key neither use for sig nor enc
		GenerateClientSecretOctetKeyFunc: secrets.GenerateOctetKey,
		GenerateAdminAPIAuthKeyFunc:      secrets.GenerateRSAKey,
		GenerateSAMLIdpSigningCertificate: func() (*config.SAMLIdpSigningCertificate, error) {
			return config.GenerateSAMLIdpSigningCertificate(commonName)
		},
	}
	incoming, err := updateInstruction.ApplyTo(updateInstructionContext, original)
	if err != nil {
		return nil, err
	}

	err = d.validate(ctx, original, incoming, fc)
	if err != nil {
		return nil, err
	}

	updatedYAML, err := yaml.Marshal(incoming)
	if err != nil {
		return nil, err
	}

	newResrc := *resrc
	newResrc.Data = updatedYAML
	return &newResrc, nil
}

func (d AuthgearSecretYAMLDescriptor) validate(ctx context.Context, original *config.SecretConfig, incoming *config.SecretConfig, fc *config.FeatureConfig) error {
	validationCtx := &validation.Context{}

	featureConfigErr := func() error {
		incomingFCError := d.validateBasedOnFeatureConfig(incoming, fc)
		incomingAggregatedError, ok := incomingFCError.(*validation.AggregatedError)
		if incomingFCError == nil || !ok {
			return incomingFCError
		}
		// https://github.com/authgear/authgear-server/commit/888e57b4b6fa9de7cd5786111cdc5cc244a85ac0
		// If the original config has some feature config error, we allow the user
		// to save the config without correcting them. This is for the case that
		// the app is downgraded from a higher plan.
		originalFCError := d.validateBasedOnFeatureConfig(original, fc)
		originalAggregatedError, ok := originalFCError.(*validation.AggregatedError)
		if originalFCError == nil || !ok {
			return incomingFCError
		}

		aggregatedError := incomingAggregatedError.Subtract(originalAggregatedError)
		return aggregatedError
	}()

	validationCtx.AddError(featureConfigErr)

	return validationCtx.Error(fmt.Sprintf("invalid %v", AuthgearSecretYAML))
}

func (d AuthgearSecretYAMLDescriptor) validateBasedOnFeatureConfig(secretConfig *config.SecretConfig, fc *config.FeatureConfig) error {
	validationCtx := &validation.Context{}

	if *fc.Messaging.CustomSMTPDisabled {
		if _, _, ok := secretConfig.Lookup(config.SMTPServerCredentialsKey); ok {
			validationCtx.EmitErrorMessage("custom smtp is not allowed")
		}
	}

	return validationCtx.Error("features are limited by feature config")
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

func (d AuthgearFeatureYAMLDescriptor) ViewResources(ctx context.Context, resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
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

	switch rawView.(type) {
	case resource.AppFileView:
		return app()
	case resource.EffectiveFileView:
		return app()
	case resource.EffectiveResourceView:
		return d.viewEffectiveResource(ctx, resources)
	case resource.ValidateResourceView:
		return d.viewEffectiveResource(ctx, resources)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (d AuthgearFeatureYAMLDescriptor) viewEffectiveResource(ctx context.Context, resources []resource.ResourceFile) (interface{}, error) {
	var cfgs []*config.FeatureConfig
	for _, layer := range resources {
		cfg, err := config.ParseFeatureConfigWithoutDefaults(ctx, layer.Data)
		if err != nil {
			return nil, fmt.Errorf("malformed feature config: %w", err)
		}
		cfgs = append(cfgs, cfg)
	}

	mergedConfig := &config.FeatureConfig{}
	for _, cfg := range cfgs {
		mergedConfig = mergedConfig.Merge(cfg)
	}
	mergedYAML, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return nil, err
	}

	featureConfig, err := config.ParseFeatureConfig(ctx, mergedYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot parse merged feature config: %w", err)
	}
	return featureConfig, nil
}

func (d AuthgearFeatureYAMLDescriptor) UpdateResource(_ context.Context, _ []resource.ResourceFile, resrc *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	return nil, fmt.Errorf("cannot update '%v'", AuthgearFeatureYAML)
}

var FeatureConfig = resource.RegisterResource(AuthgearFeatureYAMLDescriptor{})

func ParseAuthgearSecretsYAMLUpdateInstructions(data []byte) (*config.SecretConfigUpdateInstruction, error) {
	var out config.SecretConfigUpdateInstruction
	err := json.Unmarshal(data, &out)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret config update instruction: %w", err)
	}

	return &out, nil
}
