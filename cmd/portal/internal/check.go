package internal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

type CheckConfigSourcesOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	AppIDs         []string
}

func CheckConfigSources(ctx context.Context, opts *CheckConfigSourcesOptions) error {
	db := openDB(opts.DatabaseURL, opts.DatabaseSchema)

	configSources, err := selectConfigSources(ctx, db, opts.AppIDs)
	if err != nil {
		log.Fatalf("failed to connect db: %s", err)
	}

	var errs []error
	invalid := 0
	for _, c := range configSources {
		log.Printf("checking %q...", c.AppID)
		if err := checkConfigSource(c); err != nil {
			err = fmt.Errorf("invalid config source for %q: %s", c.AppID, err)
			log.Println(err.Error())
			errs = append(errs, err)
			invalid++
			continue
		}
	}
	log.Printf("checked %d; %d invalid.", len(configSources), invalid)
	return errors.Join(errs...)
}

func checkConfigSource(cs *ConfigSource) error {
	mainYAML, err := base64.StdEncoding.DecodeString(cs.Data[configsource.AuthgearYAML])
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	appConfig, err := config.Parse(mainYAML)
	if err != nil {
		return fmt.Errorf("invalid app config: %w", err)
	}

	secretsYAML, err := base64.StdEncoding.DecodeString(cs.Data[configsource.AuthgearSecretYAML])
	if err != nil {
		return fmt.Errorf("failed decode authgear.secrets.yaml: %w", err)
	}

	secretsConfig, err := config.ParseSecret(secretsYAML)
	if err != nil {
		return fmt.Errorf("invalid secrets config: %w", err)
	}

	// FIXME: For production deployment, there should be a base secrets config
	// provided by server operator, so that some required secrets are not set
	// in the config source (e.g. db/redis).
	// We simulate it with dummy value for now.

	baseSecrets := &config.SecretConfig{
		Secrets: []config.SecretItem{
			{Key: config.DatabaseCredentialsKey, Data: &config.DatabaseCredentials{}},
			{Key: config.RedisCredentialsKey, Data: &config.RedisCredentials{}},
		},
	}

	if err := baseSecrets.Overlay(secretsConfig).Validate(appConfig); err != nil {
		return fmt.Errorf("invalid secrets config: %w", err)
	}

	if _, ok := cs.Data[configsource.AuthgearFeatureYAML]; ok {
		featuresYAML, err := base64.StdEncoding.DecodeString(cs.Data[configsource.AuthgearFeatureYAML])
		if err != nil {
			return fmt.Errorf("failed decode authgear.features.yaml: %w", err)
		}

		_, err = config.ParseFeatureConfig(featuresYAML)
		if err != nil {
			return fmt.Errorf("invalid features config: %w", err)
		}
	}

	return nil
}
